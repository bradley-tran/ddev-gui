# Install WSL2 with Ubuntu
Run the following in PowerShell terminal:

`wsl --install`

You’ll probably need to reboot.

If you had previously installed WSL, update it:


`wsl --update`

Then, install an Ubuntu distro named "DDEV":


`wsl --install Ubuntu --name DDEV`

Verify that you now have an Ubuntu distro set as default:


`wsl -l -v`

```sh
  NAME                   STATE           VERSION
* DDEV                   Stopped         2
```
