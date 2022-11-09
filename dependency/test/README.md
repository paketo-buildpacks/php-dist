To test locally:

```shell
# Assume $output_dir is the output from the compilation step, with a tarball and a checksum in it.
# Note that the wildcard is not quoted, to allow globbing

# Passing
$ ./test.sh \
  --tarballPath ${output_dir}/*.tgz \
  --expectedVersion 8.1.12
Outside image: tarball_path=</path/to/output_dir>php_8.1.12_linux_x64_bionic_262d9c4e.tgz
Outside image: expectedVersion=8.1.12
# Docker compilation output
Inside image: tarballPath=/tarball_path/php_8.1.12_linux_x64_bionic_262d9c4e.tgz
Inside image: expectedVersion=8.1.12
All tests passed!

# Failing
$ /tmp/test/test.sh \
  --tarballPath ${output_dir}/*.tgz \
  --expectedVersion 999.999.999
tarballPath=/tmp/output_dir/php_8.1.12_linux_x64_bionic_262d9c4e.tgz
expectedVersion=999.999.999
Version 8.1.12 does not match expected version 999.999.999
```


## Note
You may see a large output block that looks like the following.
I have no idea why this happens, but it also happened in the `dep-server` test workflow.

https://github.com/paketo-buildpacks/dep-server/actions/runs/3385269156/jobs/5623220597

```shell
MIB search path: /github/home/.snmp/mibs:/usr/share/snmp/mibs:/usr/share/snmp/mibs/iana:/usr/share/snmp/mibs/ietf:/usr/share/mibs/site:/usr/share/snmp/mibs:/usr/share/mibs/iana:/usr/share/mibs/ietf:/usr/share/mibs/netsnmp
Cannot find module (SNMPv2-MIB): At line 0 in (none)
Cannot find module (IF-MIB): At line 0 in (none)
Cannot find module (IP-MIB): At line 0 in (none)
Cannot find module (TCP-MIB): At line 0 in (none)
Cannot find module (UDP-MIB): At line 0 in (none)
Cannot find module (HOST-RESOURCES-MIB): At line 0 in (none)
Cannot find module (NOTIFICATION-LOG-MIB): At line 0 in (none)
Cannot find module (DISMAN-EVENT-MIB): At line 0 in (none)
Cannot find module (DISMAN-SCHEDULE-MIB): At line 0 in (none)
Cannot find module (HOST-RESOURCES-TYPES): At line 0 in (none)
Cannot find module (MTA-MIB): At line 0 in (none)
Cannot find module (NETWORK-SERVICES-MIB): At line 0 in (none)
Cannot find module (UCD-DISKIO-MIB): At line 0 in (none)
Cannot find module (UCD-DLMOD-MIB): At line 0 in (none)
Cannot find module (LM-SENSORS-MIB): At line 0 in (none)
Cannot find module (UCD-SNMP-MIB): At line 0 in (none)
Cannot find module (UCD-DEMO-MIB): At line 0 in (none)
Cannot find module (SNMP-TARGET-MIB): At line 0 in (none)
Cannot find module (NET-SNMP-AGENT-MIB): At line 0 in (none)
Cannot find module (SNMP-MPD-MIB): At line 0 in (none)
Cannot find module (SNMP-USER-BASED-SM-MIB): At line 0 in (none)
Cannot find module (SNMP-FRAMEWORK-MIB): At line 0 in (none)
Cannot find module (SNMP-VIEW-BASED-ACM-MIB): At line 0 in (none)
Cannot find module (SNMP-COMMUNITY-MIB): At line 0 in (none)
Cannot find module (IPV6-ICMP-MIB): At line 0 in (none)
Cannot find module (IPV6-MIB): At line 0 in (none)
Cannot find module (IPV6-TCP-MIB): At line 0 in (none)
Cannot find module (IPV6-UDP-MIB): At line 0 in (none)
Cannot find module (IP-FORWARD-MIB): At line 0 in (none)
Cannot find module (NET-SNMP-PASS-MIB): At line 0 in (none)
Cannot find module (NET-SNMP-EXTEND-MIB): At line 0 in (none)
Cannot find module (SNMP-NOTIFICATION-MIB): At line 0 in (none)
Cannot find module (SNMPv2-TM): At line 0 in (none)
Cannot find module (NET-SNMP-VACM-MIB): At line 0 in (none)
```