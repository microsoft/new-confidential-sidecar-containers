// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

//go:build linux
// +build linux

package main

import (
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/Microsoft/confidential-sidecar-containers/pkg/attest"
	"github.com/Microsoft/confidential-sidecar-containers/pkg/common"
	"github.com/Microsoft/confidential-sidecar-containers/pkg/skr"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"

	"golang.org/x/crypto/hkdf"
)

// Test dependencies
var (
	_azmountRun                    = azmountRun
	_containerMountAzureFilesystem = containerMountAzureFilesystem
	_cryptsetupOpen                = cryptsetupOpen
	_veritysetupOpen               = veritysetupOpen
	ioutilWriteFile                = ioutil.WriteFile
	osGetenv                       = os.Getenv
	osMkdirAll                     = os.MkdirAll
	osRemoveAll                    = os.RemoveAll
	osStat                         = os.Stat
	osCreate                       = os.Create
	timeSleep                      = time.Sleep
	unixMount                      = unix.Mount
)

var (
	Identity              common.Identity
	CertState             attest.CertState
	EncodedUvmInformation common.UvmInformation
	// for testing encrypted filesystems without releasing secrets from
	// AKV allowTestingWithRawKey needs to be set to true and a raw key
	// needs to have been provided. Default mode is that such testing is
	// disabled.
	allowTestingWithRawKey = false
)

// Constant
// offset of "Roothash:"
const ROOTHASH_OFFSET int = 9
// length of roothash
const ROOTHASH_LENGTH int = 64

// azmountRun starts azmount with the specified arguments, and leaves it running
// in the background.
func azmountRun(imageLocalFolder string, azureImageUrl string, azureImageUrlPrivate bool, azmountLogFile string, cacheBlockSize string, numBlocks string, readWrite bool) error {
	identityJson, err := json.Marshal(Identity)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal identity")
	}

	encodedIdentity := base64.StdEncoding.EncodeToString(identityJson)

	logrus.Debugf("Starting azmount: -mountpoint %s -url %s -private %s -logfile %s -blocksize %s KB -numblock %s -readWrite %s", imageLocalFolder, azureImageUrl, strconv.FormatBool(azureImageUrlPrivate), azmountLogFile, cacheBlockSize, numBlocks, strconv.FormatBool(readWrite))
	cmd := exec.Command("/bin/azmount", "-mountpoint", imageLocalFolder, "-url", azureImageUrl, "-private", strconv.FormatBool(azureImageUrlPrivate), "-identity", encodedIdentity, "-logfile", azmountLogFile, "-blocksize", cacheBlockSize, "-numblocks", numBlocks, "-readWrite", strconv.FormatBool(readWrite))
	if err := cmd.Start(); err != nil {
		return errors.Wrapf(err, "azmount failed to start")
	}
	logrus.Infof("azmount running...")
	return nil
}

// cryptsetupCommand runs cryptsetup with the provided arguments
func cryptsetupCommand(args []string) error {
	// --debug and -v are used to increase the information printed by
	// cryptsetup. By default, it doesn't print much information, which makes it
	// hard to debug it when there are problems.
	logrus.Debugf("Executing cryptsetup with args: %s", append([]string{"--debug", "-v"}, args...))
	cmd := exec.Command("cryptsetup", append([]string{"--debug", "-v"}, args...)...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "failed to execute cryptsetup: %s", string(output))
	}
	return nil
}

// cryptsetupOpen runs "cryptsetup luksOpen" with the right arguments.
func cryptsetupOpen(source string, deviceName string, keyFilePath string) error {
	openArgs := []string{
		// Open device with the key passed to luksFormat
		"luksOpen", source, deviceName, "--key-file", keyFilePath,
		// Don't use a journal to increase performance
		"--integrity-no-journal",
		"--persistent"}

	return cryptsetupCommand(openArgs)
}

// veritysetupCommand runs veritysetup with the provided arguments
func veritysetupCommand(args []string) (string, error) {
	cmd := exec.Command("veritysetup", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "failed to execute veritysetup: %s", string(output))
	}
	return nil
}

// veritysetupOpen runs "veritysetup open" with right arguments
func veritysetupOpen(dataDevicePath string, dmVerityName string, hashDevicePath string, rootHash string) (string, error) {
	openArgs := []string{
		"open", dataDevicePath, dmVerityName, hashDevicePath, rootHash}
	return veritysetupCommand(openArgs)
}

// store root hash for future verification
func storeRootHash(rootHash string, mountPoint string, index int) error {
	rootHashPath, err := filepath.Abs(filepath.Join(mountPoint, fmt.Sprintf("../.dm-verity-root-hash-%d", index)))
	if err != nil {
		return errors.Wrapf(err, "failed to resolve absolute path of root hash file")
	}
	rootHashFile, err := osCreate(rootHashPath)
	if err != nil {
		return errors.Wrapf(err, "failed to create root hash file")
	}
	defer func(){
		err := rootHashFile.Close()
		if err != nil {
			logrus.WithError(err).Debugf("failed to close root hash file: %s", rootHashPath)
		} else {
			logrus.Debugf("Close root hash file: %s", rootHashPath)
		}
	}()
	_, err = rootHashFile.WriteString(rootHash)
	if err != nil {
		return errors.Wrapf(err, "failed to write root hash")
	}
	return nil
}

func mountAzureFile(tempDir string, index int, azureImageUrl string, azureImageUrlPrivate bool, cacheBlockSize string, numBlocks string, readWrite bool) (string, error) {

	imageLocalFolder := filepath.Join(tempDir, fmt.Sprintf("%d", index))
	if err := osMkdirAll(imageLocalFolder, 0755); err != nil {
		return "", errors.Wrapf(err, "mkdir failed: %s", imageLocalFolder)
	}

	// Location in the UVM of the encrypted filesystem image.
	imageLocalFile := filepath.Join(imageLocalFolder, "data")
	logrus.Debugf("Location in the UVM of the encrypted filesystem image %s", imageLocalFile)

	// Location of log file generated by azmount
	azmountLogFile := filepath.Join(tempDir, fmt.Sprintf("log-%d.txt", index))
	logrus.Debugf("Location of log file generated by azmount %s", azmountLogFile)

	// Any program that sets up a FUSE filesystem becomes a server that listens
	// to requests from the kernel, and it gets stuck in the loop that serves
	// requests, so it is needed to run it in a different process so that the
	// execution can continue in this one.
	_azmountRun(imageLocalFolder, azureImageUrl, azureImageUrlPrivate, azmountLogFile, cacheBlockSize, numBlocks, readWrite)

	// Wait until the file is available
	count := 0
	for {
		_, err := osStat(imageLocalFile)
		if err == nil {
			// Found
			break
		}
		// Timeout after 10 seconds
		count++
		if count == 1000 {
			return "", errors.Wrapf(err, "timed out while waiting for encrypted filesystem image")
		}
		timeSleep(60 * time.Millisecond)
	}
	logrus.Debugf("Encrypted file system image found: %s", imageLocalFile)

	return imageLocalFile, nil
}

// rawRemoteFilesystemKey sets up the key file path using the raw key passed
func rawRemoteFilesystemKey(tempDir string, rawKeyHexString string) (keyFilePath string, err error) {
	keyFilePath = filepath.Join(tempDir, "keyfile")

	keyBytes := make([]byte, 64)
	keyBytes, err = hex.DecodeString(rawKeyHexString)
	if err != nil {
		return "", errors.Wrapf(err, "failed to decode raw key")
	}

	// dm-crypt expects a key file, so create a key file using the key released in
	// previous step
	err = ioutilWriteFile(keyFilePath, keyBytes, 0644)
	if err != nil {
		return "", errors.Wrapf(err, "failed to create keyfile: %s", keyFilePath)
	}

	return keyFilePath, nil
}

// releaseRemoteFilesystemKey releases the key identified by keyBlob from AKV
//
// 1) Retrieve encoded  security policy by reading the environment variable
//
// 2) Perform secure key release
//
// 3) Prepare the key file path using the released key
func releaseRemoteFilesystemKey(tempDir string, keyDerivationBlob common.KeyDerivationBlob, keyBlob common.KeyBlob) (keyFilePath string, err error) {
	keyFilePath = filepath.Join(tempDir, "keyfile")

	// 2) release key identified by keyBlob using encoded security policy and certfetcher (contained in CertState object)
	//    certfetcher is required for validating the attestation report against the cert
	//    chain of the chip identified in the attestation report
	logrus.Info("Performing Secure Key Release...")
	jwKey, err := skr.SecureKeyRelease(Identity, CertState, keyBlob, EncodedUvmInformation)
	if err != nil {
		return "", errors.Wrapf(err, "failed to release key: %v", keyBlob)
	}
	logrus.Debugf("Key Type: %s", jwKey.KeyType())

	octetKeyBytes := make([]byte, 32)
	var rawKey interface{}
	err = jwKey.Raw(&rawKey)
	if err != nil {
		return "", errors.Wrapf(err, "failed to extract raw key")
	}

	if jwKey.KeyType() == "oct" {
		rawOctetKeyBytes, ok := rawKey.([]byte)
		if !ok || len(rawOctetKeyBytes) != 32 {
			return "", errors.Wrapf(err, "expected 32-byte octet key")
		}
		octetKeyBytes = rawOctetKeyBytes
	} else if jwKey.KeyType() == "RSA" {
		rawKey, ok := rawKey.(*rsa.PrivateKey)
		if !ok {
			return "", errors.Wrapf(err, "expected RSA key")
		}
		// use sha256 as hashing function for HKDF
		hash := sha256.New
		logrus.Trace("Using SHA256 as hashing function for HKDF")

		// public salt and label
		var labelString string
		if keyDerivationBlob.Label != "" {
			labelString = keyDerivationBlob.Label
		} else {
			labelString = "Symmetric Encryption Key"
		}
		logrus.Debugf("Key Derivation Label: %s", labelString)

		// decode public salt hexstring
		salt, err := hex.DecodeString(keyDerivationBlob.Salt)
		if err != nil {
			return "", errors.Wrapf(err, "failed to decode Key Derivation Salt hexstring")
		}

		// setup derivation function using secret D exponent, salt, and label
		logrus.Trace("Setup symmetric key derivation function using HKDF with secret D exponent, salt, and label...")
		hkdf := hkdf.New(hash, rawKey.D.Bytes(), salt, []byte(labelString))

		// derive key
		logrus.Trace("Deriving symmetric key...")
		if _, err := io.ReadFull(hkdf, octetKeyBytes); err != nil {
			return "", errors.Wrapf(err, "failed to derive oct key")
		}

		logrus.Debugf("Symmetric key %s (salt: %s label: %s)", hex.EncodeToString(octetKeyBytes), keyDerivationBlob.Salt, labelString)
	} else {
		return "", errors.Wrapf(err, "key type %s not supported", jwKey.KeyType())
	}

	// 3) dm-crypt expects a key file, so create a key file using the key released in
	//    previous step
	logrus.Debugf("Creating keyfile: %s", keyFilePath)
	err = ioutilWriteFile(keyFilePath, octetKeyBytes, 0644)
	if err != nil {
		return "", errors.Wrapf(err, "failed to create keyfile: %s", keyFilePath)
	}

	return keyFilePath, nil
}

// containerMountAzureFilesystem mounts a remote filesystems specified in the
// policy of a given container.
//
//  1. Get the actual filesystem image. This is done by starting a new azmount
//     process. The file is then exposed at “/[tempDir]/[index]/data“ and the
//     log of azmount is saved to “/[tempDir]/log-[index].txt“.
//
//  2. Obtain keyfile. This is hardcoded at the moment and needs to be replaced
//     by the actual code that gets the key. It is saved to a temporary file so
//     that it can be passed to cryptsetup. It can be removed afterwards.
//
//  3. Open encrypted filesystem with cryptsetup. The result is a block device in
//     “/dev/mapper/remote-crypt-[filesystem-index]“.
//
//  4. Config dm-verity on /dev/mapper/remote-crypt-[filesystem-index]
//
//  5. Mount block device as a read-only filesystem.
//
//  6. Create a symlink to the filesystem in the path shared between the UVM and
//     the container.
func containerMountAzureFilesystem(tempDir string, index int, fs AzureFilesystem) (err error) {

	cacheBlockSize := "512"
	numBlocks := "32"

	// Filesystem cannot be both writable and dm-verity protected
	if fs.ReadWrite && fs.DmVerity.Enable {
		logrus.Fatalf("Dm-verity protected file system is not writable!")
	}
	// get dataTempDir and hashTempDir
	dataTempDir := filepath.Join(tempDir, "data")
	hashTempDir := filepath.Join(tempDir, "hash")
	var hashLocalFile string

	// 1) Mount remote image
	logrus.Debugf("Mounting remote image %s", fs.AzureUrl)
	dataLocalFile, err := mountAzureFile(dataTempDir, index, fs.AzureUrl, fs.AzureUrlPrivate, cacheBlockSize, numBlocks, fs.ReadWrite)
	if err != nil {
		return errors.Wrapf(err, "failed to mount remote file: %s", fs.AzureUrl)
	}
	// mount hash device if dm-verity is set true
	if fs.DmVerity.Enable == true {
		logrus.Debugf("Mounting remote hash device %s", fs.DmVerity.HashUrl)
		hashLocalFile, err = mountAzureFile(hashTempDir, index, fs.DmVerity.HashUrl, fs.AzureUrlPrivate, cacheBlockSize, numBlocks, fs.ReadWrite)
		if err != nil {
			return errors.Wrapf(err, "failed to mount remote hashDevice: %s", fs.DmVerity.HashUrl)
		}
	}

	// 2) Obtain keyfile
	logrus.Infof("Obtaining keyfile...")
	var keyFilePath string
	if fs.KeyBlob.KID != "" {
		keyFilePath, err = releaseRemoteFilesystemKey(tempDir, fs.KeyDerivationBlob, fs.KeyBlob)
		if err != nil {
			return errors.Wrapf(err, "failed to obtain keyfile %s", fs.KeyBlob.KID)
		}
	} else if allowTestingWithRawKey {
		keyFilePath, err = rawRemoteFilesystemKey(tempDir, fs.RawKeyHexString)
		if err != nil {
			return errors.Wrapf(err, "failed to obtain keyfile %s", fs.RawKeyHexString)
		}
	}

	defer func() {
		// Delete keyfile on exit
		if inErr := osRemoveAll(keyFilePath); inErr != nil {
			logrus.WithError(inErr).Debugf("failed to delete keyfile: %s", keyFilePath)
		} else {
			logrus.Debugf("Deleted keyfile: %s", keyFilePath)
		}
	}()

	// 3) Open encrypted filesystem with veritysetup if dm-verity is set true. 
	// The result is a block device in /dev/mapper/remote-verity-[filesystem-index].
	var verityDeviceName string
	var verityDevicePath string
	// open verity device	
	if fs.DmVerity.Enable == true {
		verityDeviceName = fmt.Sprintf("remote-verity-%d", index)
		verityDevicePath = "/dev/mapper/" + verityDeviceName
		err = _veritysetupOpen(dataLocalFile, verityDeviceName, hashLocalFile, fs.DmVerity.RootHash)
		if err != nil {
			return errors.Wrapf(err, "Fail to open dm-verity device")
		}
		// store root hash for future verification
		err = storeRootHash(fs.DmVerity.RootHash, fs.MountPoint, index)
		if err != nil {
			return errors.Wrapf(err, "Failed to store root hash as a file")
		}
		logrus.Infof("Successfully open dm-verity device")
	}

	// 4) Open encrypted filesystem with cryptsetup. The result is a block
	// device in /dev/mapper/remote-crypt-[filesystem-index] so that it is
	// unique from all other filesystems.
	var deviceName = fmt.Sprintf("remote-crypt-%d", index)
	var deviceNamePath = "/dev/mapper/" + deviceName

	logrus.Debugf("Opening device at: %s", deviceNamePath)
	// read from dm-verity device
	if fs.DmVerity.Enable == true {
		err = _cryptsetupOpen(verityDevicePath, deviceName, keyFilePath)
		if err != nil {
			return errors.Wrapf(err, "luksOpen failed: %s", deviceName)
		}
	} else {
		// no dm-verity
		err = _cryptsetupOpen(dataLocalFile, deviceName, keyFilePath)
		if err != nil {
			return errors.Wrapf(err, "luksOpen failed: %s", deviceName)
		}
		logrus.Debugf("Device opened: %s", deviceName)
	}

	// 5) Mount block device as a read-only filesystem.
	tempMountFolder, err := filepath.Abs(filepath.Join(fs.MountPoint, fmt.Sprintf("../.filesystem-%d", index)))
	if err != nil {
		return errors.Wrapf(err, "failed to resolve absolute path of mount point %s for filesystem-%d", fs.MountPoint, index)
	}

	logrus.Debugf("Mounting filesystem-%d to: %s", index, tempMountFolder)

	var flags uintptr
	var data string
	if !fs.ReadWrite {
		flags = unix.MS_RDONLY
		data = "noload"
	}

	logrus.Debugf("Creating mount folder: %s", tempMountFolder)
	if err := osMkdirAll(tempMountFolder, 0755); err != nil {
		return errors.Wrapf(err, "mkdir failed: %s", tempMountFolder)
	}

	logrus.Debugf("Mounting filesystem %s to mount folder %s", deviceNamePath, tempMountFolder)
	if err := unixMount(deviceNamePath, tempMountFolder, "ext4", flags, data); err != nil {
		return errors.Wrapf(err, "failed to mount filesystem: %s", deviceNamePath)
	}

	// 6) Create a symlink to the folder where the filesystem is mounted.
	destPath := fs.MountPoint
	logrus.Debugf("Creating symlink for filesystem-%d to: %s", index, destPath)

	if err := os.Symlink(fmt.Sprintf(".filesystem-%d", index), destPath); err != nil {
		return errors.Wrapf(err, "failed to symlink filesystem-%d: %s", index, destPath)
	}

	return nil
}

func MountAzureFilesystems(tempDir string, info RemoteFilesystemsInformation) (err error) {

	Identity = info.AzureInfo.Identity

	// Retrieve the incoming encoded security policy, cert and uvm endorsement
	EncodedUvmInformation, err = common.GetUvmInformation()
	if err != nil {
		logrus.Infof("Failed to extract UVM_* environment variables: %s", err.Error())
	}

	if common.ThimCertsAbsent(&EncodedUvmInformation.InitialCerts) {
		logrus.Infof("ThimCerts is absent, retrieving THIMCerts from %s.", info.AzureInfo.CertFetcher.Endpoint)
		thimCerts, err := info.AzureInfo.CertFetcher.GetThimCerts(info.AzureInfo.CertFetcher.Endpoint)
		if err != nil {
			logrus.Fatalf("Failed to retrieve thim certs: %s", err.Error())
		}
		EncodedUvmInformation.InitialCerts = *thimCerts
	}

	logrus.Debugf("EncodedUvmInformation.InitialCerts.Tcbm: %s\n", EncodedUvmInformation.InitialCerts.Tcbm)
	thimTcbm, err := strconv.ParseUint(EncodedUvmInformation.InitialCerts.Tcbm, 16, 64)
	if err != nil {
		return errors.Wrapf(err, "failed to parse THIM TCBM")
	}

	CertState = attest.CertState{
		CertFetcher: info.AzureInfo.CertFetcher,
		Tcbm:        thimTcbm,
	}

	for i, fs := range info.AzureFilesystems {
		logrus.Infof("Mounting Azure Storage blob %d...", i)

		err = _containerMountAzureFilesystem(tempDir, i, fs)
		if err != nil {
			return errors.Wrapf(err, "failed to mount filesystem index %d", i)
		}
	}

	return nil
}
