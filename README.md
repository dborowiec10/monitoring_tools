# DL-GPU-Energy-Project-Tools
Any tools related to the project, i.e. for data extraction, analysis



## disabling intel pstate driver:
add the following to the /etc/default/grub file on line `GRUB_CMDLINE_LINUX_DEFAULT`:
```
intel_pstate=disable acpi=force
```

like so:
```
GRUB_CMDLINE_LINUX_DEFAULT="intel_pstate=disable acpi=force"
```

## available frequencies:
3.60 GHz
3.40 GHz
3.30 GHz
3.10 GHz
2.90 GHz
2.70 GHz
2.60 GHz
2.40 GHz
2.20 GHz
2.10 GHz
1.90 GHz
1.70 GHz
1.50 GHz
1.40 GHz
1.20 GHz