package policy

import future.keywords.every
import future.keywords.in

api_version := "0.10.0"
framework_version := "0.2.3"

fragments := [
  {
    "feed": "mcr.microsoft.com/aci/aci-cc-infra-fragment",
    "includes": [
      "containers",
      "fragments"
    ],
    "issuer": "did:x509:0:sha256:I__iuL25oXEVFdTP_aBLx_eT1RPHbCQ_ECBQfYZpt9s::eku:1.3.6.1.4.1.311.76.59.1.3",
    "minimum_svn": "1"
  }
]

containers := [{"allow_elevated":false,"allow_stdio_access":true,"capabilities":{"ambient":[],"bounding":["CAP_AUDIT_WRITE","CAP_CHOWN","CAP_DAC_OVERRIDE","CAP_FOWNER","CAP_FSETID","CAP_KILL","CAP_MKNOD","CAP_NET_BIND_SERVICE","CAP_NET_RAW","CAP_SETFCAP","CAP_SETGID","CAP_SETPCAP","CAP_SETUID","CAP_SYS_CHROOT"],"effective":["CAP_AUDIT_WRITE","CAP_CHOWN","CAP_DAC_OVERRIDE","CAP_FOWNER","CAP_FSETID","CAP_KILL","CAP_MKNOD","CAP_NET_BIND_SERVICE","CAP_NET_RAW","CAP_SETFCAP","CAP_SETGID","CAP_SETPCAP","CAP_SETUID","CAP_SYS_CHROOT"],"inheritable":[],"permitted":["CAP_AUDIT_WRITE","CAP_CHOWN","CAP_DAC_OVERRIDE","CAP_FOWNER","CAP_FSETID","CAP_KILL","CAP_MKNOD","CAP_NET_BIND_SERVICE","CAP_NET_RAW","CAP_SETFCAP","CAP_SETGID","CAP_SETPCAP","CAP_SETUID","CAP_SYS_CHROOT"]},"command":["python3","primary.py"],"env_rules":[{"pattern":"ENCFS_MOUNT=/mnt/remote","required":false,"strategy":"string"},{"pattern":"PATH=/usr/local/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin","required":false,"strategy":"string"},{"pattern":"LANG=C.UTF-8","required":false,"strategy":"string"},{"pattern":"GPG_KEY=7169605F62C751356D054A26A821E680E5FA6305","required":false,"strategy":"string"},{"pattern":"PYTHON_VERSION=3.12.4","required":false,"strategy":"string"},{"pattern":"PYTHON_PIP_VERSION=24.0","required":false,"strategy":"string"},{"pattern":"PYTHON_GET_PIP_URL=https://github.com/pypa/get-pip/raw/dbf0c85f76fb6e1ab42aa672ffca6f0a675d9ee4/public/get-pip.py","required":false,"strategy":"string"},{"pattern":"PYTHON_GET_PIP_SHA256=dfe9fd5c28dc98b5ac17979a953ea550cec37ae1b47a5116007395bfacff2ab9","required":false,"strategy":"string"},{"pattern":"TERM=xterm","required":false,"strategy":"string"},{"pattern":"(?i)(FABRIC)_.+=.+","required":false,"strategy":"re2"},{"pattern":"HOSTNAME=.+","required":false,"strategy":"re2"},{"pattern":"T(E)?MP=.+","required":false,"strategy":"re2"},{"pattern":"FabricPackageFileName=.+","required":false,"strategy":"re2"},{"pattern":"HostedServiceName=.+","required":false,"strategy":"re2"},{"pattern":"IDENTITY_API_VERSION=.+","required":false,"strategy":"re2"},{"pattern":"IDENTITY_HEADER=.+","required":false,"strategy":"re2"},{"pattern":"IDENTITY_SERVER_THUMBPRINT=.+","required":false,"strategy":"re2"},{"pattern":"azurecontainerinstance_restarted_by=.+","required":false,"strategy":"re2"}],"exec_processes":[{"command":["/bin/sh"],"signals":[]},{"command":["/bin/bash"],"signals":[]}],"id":"confidentialsidecars.azurecr.io/encfs/primary:954243d8-517a-4d59-bd5b-32797ef64b13","layers":["8e978a1d2dbb44e1ff09eb8edc611f716bb4307c9c1b7583302d94d34daa007f","1fe6d4527bbaefd3f85b423530ec202b737a70cfae5ecd2280b944b53375979f","df13c8a3bf4ab0cdfc66fa3539ea890a7395ffff908cbb2fa1f52b7765d58399","38b7ef46db0f0423b0ac4177431b4a13c79836526f9a1076f44fc75a679c189b","2024e6fdb1de7cbafe176b5114110d811c98dbdbfbee65defb92bedcbea80b15","e3695f598fbafe8d204b9d4e5df5f253e4f33c7b580bf326eee1eb2d98b8bb3e","8a785382dfdd335b53cb12a9d165154d8cb967186ecf49e9b7eefb5a0dfc2b7a","4e84e0ae315a51b10a49d80bd016762e5b400779bb21eda38ff090bfe5ce9950","6d26f207cb6fe798bba1adfbabfe0d44ba1e0ce1dbbdb0d7e77ad66f70ef21ba","c6f2575a56e761d73655dc9a5178a060b2a255eed73ffce5df7f3a3907ac4d81","e2fab1696cf8298492f68c35880b281c9d0e16383fcfac2ee835a93f1d02ac85"],"mounts":[{"destination":"/mnt/remote","options":["rbind","rshared","rw"],"source":"sandbox:///tmp/atlas/emptydir/.+","type":"bind"},{"destination":"/etc/resolv.conf","options":["rbind","rshared","rw"],"source":"sandbox:///tmp/atlas/resolvconf/.+","type":"bind"}],"name":"primary","no_new_privileges":false,"seccomp_profile_sha256":"","signals":[],"user":{"group_idnames":[{"pattern":"","strategy":"any"}],"umask":"0022","user_idname":{"pattern":"","strategy":"any"}},"working_dir":"/usr/src/app"},{"allow_elevated":true,"allow_stdio_access":true,"capabilities":{"ambient":[],"bounding":["CAP_AUDIT_CONTROL","CAP_AUDIT_READ","CAP_AUDIT_WRITE","CAP_BLOCK_SUSPEND","CAP_BPF","CAP_CHECKPOINT_RESTORE","CAP_CHOWN","CAP_DAC_OVERRIDE","CAP_DAC_READ_SEARCH","CAP_FOWNER","CAP_FSETID","CAP_IPC_LOCK","CAP_IPC_OWNER","CAP_KILL","CAP_LEASE","CAP_LINUX_IMMUTABLE","CAP_MAC_ADMIN","CAP_MAC_OVERRIDE","CAP_MKNOD","CAP_NET_ADMIN","CAP_NET_BIND_SERVICE","CAP_NET_BROADCAST","CAP_NET_RAW","CAP_PERFMON","CAP_SETFCAP","CAP_SETGID","CAP_SETPCAP","CAP_SETUID","CAP_SYSLOG","CAP_SYS_ADMIN","CAP_SYS_BOOT","CAP_SYS_CHROOT","CAP_SYS_MODULE","CAP_SYS_NICE","CAP_SYS_PACCT","CAP_SYS_PTRACE","CAP_SYS_RAWIO","CAP_SYS_RESOURCE","CAP_SYS_TIME","CAP_SYS_TTY_CONFIG","CAP_WAKE_ALARM"],"effective":["CAP_AUDIT_CONTROL","CAP_AUDIT_READ","CAP_AUDIT_WRITE","CAP_BLOCK_SUSPEND","CAP_BPF","CAP_CHECKPOINT_RESTORE","CAP_CHOWN","CAP_DAC_OVERRIDE","CAP_DAC_READ_SEARCH","CAP_FOWNER","CAP_FSETID","CAP_IPC_LOCK","CAP_IPC_OWNER","CAP_KILL","CAP_LEASE","CAP_LINUX_IMMUTABLE","CAP_MAC_ADMIN","CAP_MAC_OVERRIDE","CAP_MKNOD","CAP_NET_ADMIN","CAP_NET_BIND_SERVICE","CAP_NET_BROADCAST","CAP_NET_RAW","CAP_PERFMON","CAP_SETFCAP","CAP_SETGID","CAP_SETPCAP","CAP_SETUID","CAP_SYSLOG","CAP_SYS_ADMIN","CAP_SYS_BOOT","CAP_SYS_CHROOT","CAP_SYS_MODULE","CAP_SYS_NICE","CAP_SYS_PACCT","CAP_SYS_PTRACE","CAP_SYS_RAWIO","CAP_SYS_RESOURCE","CAP_SYS_TIME","CAP_SYS_TTY_CONFIG","CAP_WAKE_ALARM"],"inheritable":["CAP_AUDIT_CONTROL","CAP_AUDIT_READ","CAP_AUDIT_WRITE","CAP_BLOCK_SUSPEND","CAP_BPF","CAP_CHECKPOINT_RESTORE","CAP_CHOWN","CAP_DAC_OVERRIDE","CAP_DAC_READ_SEARCH","CAP_FOWNER","CAP_FSETID","CAP_IPC_LOCK","CAP_IPC_OWNER","CAP_KILL","CAP_LEASE","CAP_LINUX_IMMUTABLE","CAP_MAC_ADMIN","CAP_MAC_OVERRIDE","CAP_MKNOD","CAP_NET_ADMIN","CAP_NET_BIND_SERVICE","CAP_NET_BROADCAST","CAP_NET_RAW","CAP_PERFMON","CAP_SETFCAP","CAP_SETGID","CAP_SETPCAP","CAP_SETUID","CAP_SYSLOG","CAP_SYS_ADMIN","CAP_SYS_BOOT","CAP_SYS_CHROOT","CAP_SYS_MODULE","CAP_SYS_NICE","CAP_SYS_PACCT","CAP_SYS_PTRACE","CAP_SYS_RAWIO","CAP_SYS_RESOURCE","CAP_SYS_TIME","CAP_SYS_TTY_CONFIG","CAP_WAKE_ALARM"],"permitted":["CAP_AUDIT_CONTROL","CAP_AUDIT_READ","CAP_AUDIT_WRITE","CAP_BLOCK_SUSPEND","CAP_BPF","CAP_CHECKPOINT_RESTORE","CAP_CHOWN","CAP_DAC_OVERRIDE","CAP_DAC_READ_SEARCH","CAP_FOWNER","CAP_FSETID","CAP_IPC_LOCK","CAP_IPC_OWNER","CAP_KILL","CAP_LEASE","CAP_LINUX_IMMUTABLE","CAP_MAC_ADMIN","CAP_MAC_OVERRIDE","CAP_MKNOD","CAP_NET_ADMIN","CAP_NET_BIND_SERVICE","CAP_NET_BROADCAST","CAP_NET_RAW","CAP_PERFMON","CAP_SETFCAP","CAP_SETGID","CAP_SETPCAP","CAP_SETUID","CAP_SYSLOG","CAP_SYS_ADMIN","CAP_SYS_BOOT","CAP_SYS_CHROOT","CAP_SYS_MODULE","CAP_SYS_NICE","CAP_SYS_PACCT","CAP_SYS_PTRACE","CAP_SYS_RAWIO","CAP_SYS_RESOURCE","CAP_SYS_TIME","CAP_SYS_TTY_CONFIG","CAP_WAKE_ALARM"]},"command":["/encfs.sh"],"env_rules":[{"pattern":"EncfsSideCarArgs=eyJhenVyZV9maWxlc3lzdGVtcyI6IFt7Im1vdW50X3BvaW50IjogIi9tbnQvcmVtb3RlLzk1NDI0M2Q4LTUxN2EtNGQ1OS1iZDViLTMyNzk3ZWY2NGIxMy1ibG9iMSIsICJhenVyZV91cmwiOiAiaHR0cHM6Ly9jYWNpc2lkZWNhcnNzdG9yYWdlLmJsb2IuY29yZS53aW5kb3dzLm5ldC9jb250YWluZXIvOTU0MjQzZDgtNTE3YS00ZDU5LWJkNWItMzI3OTdlZjY0YjEzLWJsb2IxIiwgImF6dXJlX3VybF9wcml2YXRlIjogdHJ1ZSwgInJlYWRfd3JpdGUiOiB0cnVlLCAia2V5IjogeyJraWQiOiAiOTU0MjQzZDgtNTE3YS00ZDU5LWJkNWItMzI3OTdlZjY0YjEzLWtleSIsICJhdXRob3JpdHkiOiB7ImVuZHBvaW50IjogImNvbmZpZGVudGlhbHNpZGVjYXJzLndldS5hdHRlc3QuYXp1cmUubmV0In0sICJha3YiOiB7ImVuZHBvaW50IjogImNhY2lzaWRlY2Fycy5tYW5hZ2VkaHNtLmF6dXJlLm5ldCJ9fX1dfQ==","required":false,"strategy":"string"},{"pattern":"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin","required":false,"strategy":"string"},{"pattern":"BUILD_DIR=/go/src/github.com/microsoft/confidential-sidecar-containers","required":false,"strategy":"string"},{"pattern":"TERM=xterm","required":false,"strategy":"string"},{"pattern":"(?i)(FABRIC)_.+=.+","required":false,"strategy":"re2"},{"pattern":"HOSTNAME=.+","required":false,"strategy":"re2"},{"pattern":"T(E)?MP=.+","required":false,"strategy":"re2"},{"pattern":"FabricPackageFileName=.+","required":false,"strategy":"re2"},{"pattern":"HostedServiceName=.+","required":false,"strategy":"re2"},{"pattern":"IDENTITY_API_VERSION=.+","required":false,"strategy":"re2"},{"pattern":"IDENTITY_HEADER=.+","required":false,"strategy":"re2"},{"pattern":"IDENTITY_SERVER_THUMBPRINT=.+","required":false,"strategy":"re2"},{"pattern":"azurecontainerinstance_restarted_by=.+","required":false,"strategy":"re2"}],"exec_processes":[{"command":["/bin/sh"],"signals":[]},{"command":["/bin/bash"],"signals":[]}],"id":"confidentialsidecars.azurecr.io/encfs/sidecar:954243d8-517a-4d59-bd5b-32797ef64b13","layers":["c489fc14f658c0d3e71a94eef8bc374648b17a55ee03678c2ffc0b8676fc4ea1","51eb10a2a6400fafa1a123a46bb9ff9a2628dc5e46d3ffe6cca9e1162737102f","99c0907d4454502a5d9dc4e966f5c20e770b8f91467a8560a7f0e6565b80d3ad","5101f81b8596f285d0136f63dc2e755ea698363b6ce49dfa811f2f2b42f7fcee","76ec4e048e5ebeb0982b7b2522401d1d89a7c900c713fc24ac4cc22717a6a2fe"],"mounts":[{"destination":"/mnt/remote","options":["rbind","rshared","rw"],"source":"sandbox:///tmp/atlas/emptydir/.+","type":"bind"},{"destination":"/etc/resolv.conf","options":["rbind","rshared","rw"],"source":"sandbox:///tmp/atlas/resolvconf/.+","type":"bind"}],"name":"sidecar","no_new_privileges":false,"seccomp_profile_sha256":"","signals":[],"user":{"group_idnames":[{"pattern":"","strategy":"any"}],"umask":"0022","user_idname":{"pattern":"","strategy":"any"}},"working_dir":"/"},{"allow_elevated":false,"allow_stdio_access":true,"capabilities":{"ambient":[],"bounding":["CAP_CHOWN","CAP_DAC_OVERRIDE","CAP_FSETID","CAP_FOWNER","CAP_MKNOD","CAP_NET_RAW","CAP_SETGID","CAP_SETUID","CAP_SETFCAP","CAP_SETPCAP","CAP_NET_BIND_SERVICE","CAP_SYS_CHROOT","CAP_KILL","CAP_AUDIT_WRITE"],"effective":["CAP_CHOWN","CAP_DAC_OVERRIDE","CAP_FSETID","CAP_FOWNER","CAP_MKNOD","CAP_NET_RAW","CAP_SETGID","CAP_SETUID","CAP_SETFCAP","CAP_SETPCAP","CAP_NET_BIND_SERVICE","CAP_SYS_CHROOT","CAP_KILL","CAP_AUDIT_WRITE"],"inheritable":[],"permitted":["CAP_CHOWN","CAP_DAC_OVERRIDE","CAP_FSETID","CAP_FOWNER","CAP_MKNOD","CAP_NET_RAW","CAP_SETGID","CAP_SETUID","CAP_SETFCAP","CAP_SETPCAP","CAP_NET_BIND_SERVICE","CAP_SYS_CHROOT","CAP_KILL","CAP_AUDIT_WRITE"]},"command":["/pause"],"env_rules":[{"pattern":"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin","required":true,"strategy":"string"},{"pattern":"TERM=xterm","required":false,"strategy":"string"}],"exec_processes":[],"layers":["16b514057a06ad665f92c02863aca074fd5976c755d26bff16365299169e8415"],"mounts":[],"no_new_privileges":false,"seccomp_profile_sha256":"","signals":[],"user":{"group_idnames":[{"pattern":"","strategy":"any"}],"umask":"0022","user_idname":{"pattern":"","strategy":"any"}},"working_dir":"/"}]

allow_properties_access := true
allow_dump_stacks := true
allow_runtime_logging := true
allow_environment_variable_dropping := true
allow_unencrypted_scratch := false
allow_capability_dropping := true

mount_device := data.framework.mount_device
unmount_device := data.framework.unmount_device
mount_overlay := data.framework.mount_overlay
unmount_overlay := data.framework.unmount_overlay
create_container := data.framework.create_container
exec_in_container := data.framework.exec_in_container
exec_external := data.framework.exec_external
shutdown_container := data.framework.shutdown_container
signal_container_process := data.framework.signal_container_process
plan9_mount := data.framework.plan9_mount
plan9_unmount := data.framework.plan9_unmount
get_properties := data.framework.get_properties
dump_stacks := data.framework.dump_stacks
runtime_logging := data.framework.runtime_logging
load_fragment := data.framework.load_fragment
scratch_mount := data.framework.scratch_mount
scratch_unmount := data.framework.scratch_unmount

reason := {"errors": data.framework.errors}


