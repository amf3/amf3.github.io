---
title: "cloud-init troubleshooting"
description: A simple workflow in resolving cloud-init deployment problems.
date: 2025-03-21T16:28:54-04:00
draft: false
searchHidden: false
showtoc: true
categories: [cloud, automation, operations]
---

I previously wrote an [introduction](../../../articles/cloudinit/intro/) to cloud-init. I'd like to now follow up with a discussion on
troubleshooting. cloud-init failures on remote hosts can be challenging. Depending on the failure point, cloud-init may or may not
provide clear error indicators.  These are methods I use during provisioning issues related to cloud-init.

# Understanding cloud-init execution stages

Before continuing, let's cover some background.  cloud-init follows
[five stages](https://cloudinit.readthedocs.io/en/latest/explanation/boot.html) during boot which run sequentially. If
a stage completes, output will contain a status that can be used to verify that stage was successful.  

**Detect stage:** The init system is responsible for calling
[ds_identify](https://github.com/canonical/cloud-init/blob/main/tools/ds-identify) to determine whether cloud-init 
should run.  With systemd hosts, this is implemented as a systemd generator.

**Local stage:** Identifies local resources that are available without network access. Configures networking, which if unsuccessful
falls back to DHCP.

**Network stage:** Retrieves user-data, sets up disk partitions, and mounts the filesystem.  When complete, serial console or SSH access 
should become available.

**Config stage:** Runs configuration modules and executes commands specified in user-data.

**Final stage:** Installs packages, applies configuration management plugins like puppet or chef, and runs user or vendor defined scripts.

# Checking Stage Status

The **status** submenu from the cloud-init command provides a method of checking each stage for errors.  In this
example I intentionally mistyped a schema key name that should be `passwd` as `password`.  Output shows the failure
occurred during the init stage & provides a suggestion on how to resolve it.

```shell
$ cloud-init status --format json
...
  "extended_status": "degraded done",
  "init": {
    "errors": [],
    "finished": 6.52,
    "recoverable_errors": {
      "WARNING": [
        "cloud-config failed schema validation! You may run 'sudo cloud-init schema --system' to check the details."
      ]
    },
...
  "status": "done"
```

# Checking logs for Errors

When the issue is not obvious, there logs are available for further examination.

- /var/log/cloud-init.log  (execution details and errors)
- /var/log/cloud-init-output.log  (captured output from executed commands)
- /run/cloud-init/result.json  (summary of execution status)

Example log output from cloud-init.log indicating a schema validation failure.

```shell
2025-03-18 11:46:41,379 - schema.py[WARNING]: cloud-config failed schema validation! You may run 'sudo cloud-init schema --system' to check the details.
```

# Debugging User-Data Issues

cloud-init has a defined schema and it’s possible to validate user-data content with the **schema** submenu.
To troubleshoot a possible schema issue on a remote host where a YAML key named `passwd` was mistyped as `password`.

```shell
$ sudo cloud-init schema --system
Found cloud-config data types: user-data, vendor-data, network-config

1. user-data at /var/lib/cloud/instances/docker-demo/cloud-config.txt:
  Invalid user-data /var/lib/cloud/instances/docker-demo/cloud-config.txt
  Error: Cloud config schema errors: users.0: Additional properties are not allowed ('password' was unexpected)
…
Error: Invalid schema: user-data
```

To test changes made to user-data content prior to provisioning: `cloud-init schema -c “my_user_data_file.yaml”`.

For timeout issues in user or vendor scripts, `cloud-init analyze` will print execution times which pinpoint delays.

# Common Failure Scenarios and Fixes

A typical source of failures is from syntax errors in the user-data file.  As already mentioned, `cloud-init schema` will
show schema issues in user-data.  Manually check for typos within the values in user-data. A mistyped value is
still a string and can pass the schema validation.

Another possible issue is misconfigured network settings preventing package installation.  Ensure package mirrors are reachable
and contain the package.  The **cloud-init-output.log** file can show additional hints related to package failures.  If SSH is unavailable,
try accessing the instance over the instance's serial console.

Check for missing or incorrectly set permissions on scripts.  

Use `cloud-init analyze show` to help in identifying long-running stages.

```shell
$ cloud-init analyze show

-- Boot Record 01 --
The total time elapsed since completing an event is printed after the "@" character.
The time the event takes is printed after the "+" character.

Starting stage: init-local
|`->no cache found @00.00100s +00.00000s
|`->found local data from DataSourceNoCloud @00.00400s +00.01500s
Finished stage: (init-local) 00.28900 seconds

Starting stage: init-network
|`->restored from cache with run check: DataSourceNoCloud [seed=/dev/vda] @02.56800s +00.00100s
|`->setting up datasource @02.57600s +00.00000s
|`->reading and applying user-data @02.58000s +00.00200s
|`->reading and applying vendor-data @02.58200s +00.00200s
|`->reading and applying vendor-data2 @02.58400s +00.00000s
...
```

# Recovery and Re-Runs

Additional steps are needed after modifying user-data files on the failed instance. When cloud-init runs, output is 
cached to disk.  If the cache exists on reboot, cloud-init will not run again.  To clean cached instance data, 
run `cloud-init clean --logs` and reboot the instance.

Issues with a specific module can be exposed by using `cloud-init single`.  This could be useful when
troubleshooting user or vendor scripts.  Module names can be found with `grep "Running module" /var/log/cloud-init.log`.  

```shell
$ sudo cloud-init single --name set_passwords
Cloud-init v. 24.4.1-0ubuntu0~24.04.1 running 'single' at Fri, 21 Mar 2025 20:45:47 +0000. Up 16145.16 seconds.
```

When using the **single** submenu, it won't necessarily fix dependencies unless those are also explicitly re-triggered.  It's best
to reprovision the instance after troubleshooting any failed modules.

# Takeaways

There’s no simple fix for understanding why instance provisioning with cloud-init failed.  That’s why I’m
closing with a checklist.

* Check cloud-init status
    * Use `cloud-init status --long` (or --json) for execution state
    * Use `cloud-init analyze` for timing analysis
* Inspect logs for errors
    * /var/log/cloud-init.log: Shows errors and execution order
    * /var/log/cloud-init-output.log: contains command output
* Validate user-data input
    * `cloud-init schema` to validate syntax
    * Ensure values are correct and not only properly formatted YAML
* Check for missing dependencies or network issues
    * Ensure package mirrors are available and contain the necessary packages.
    * Verify custom scripts are executable.
* Re-run cloud-init if needed.
    * Clean logs and reset cloud-init: `cloud-init clean --logs` && reboot
    * Manually rerun a failed module: `cloud-init single -n “some_module_name”`

Happy provisioning, and follow me on [Bluesky](https://bsky.app/profile/af9.us) if you find content like this interesting.