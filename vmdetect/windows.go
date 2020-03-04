// +build windows

package vmdetect

import (
	"fmt"
	"golang.org/x/sys/windows/registry"
	"os"
	"path"
	"strings"
)

func doesRegistryKeyExist(key string) bool {

	subkeyPrefix := ""

	// Handle trailing wildcard
	if key[len(key)-1:] == "*" {
		key, subkeyPrefix = path.Split(key)
		subkeyPrefix = subkeyPrefix[:len(subkeyPrefix)-1] // remove *
	}

	firstSeparatorIndex := strings.Index(key, string(os.PathSeparator))
	keyTypeStr := key[:firstSeparatorIndex]
	keyPath := key[firstSeparatorIndex+1:]

	var keyType registry.Key
	switch keyTypeStr {
	case "HKLM":
		keyType = registry.LOCAL_MACHINE
		break
	case "HKCR":
		keyType = registry.CLASSES_ROOT
		break
	case "HKCU":
		keyType = registry.CURRENT_USER
		break
	case "HKU":
		keyType = registry.USERS
		break
	case "HKCC":
		keyType = registry.CURRENT_CONFIG
		break
	default:
		PrintError(fmt.Sprintf("Invalid keytype (%v)", keyTypeStr))
		return false
	}

	keyHandle, err := registry.OpenKey(keyType, keyPath, registry.QUERY_VALUE)

	if err != nil {
		PrintError(fmt.Sprintf("Cannot open %v : %v", key, err))
		return false
	}

	defer keyHandle.Close()

	// If a wildcard has been specified...
	if subkeyPrefix != "" {
		// ... we look for sub-keys to see if one exists
		subKeys, err := keyHandle.ReadSubKeyNames(0xFFFF)

		if err != nil {
			PrintError(err)
			return false
		}

		for _, subKeyName := range subKeys {
			if strings.HasPrefix(subKeyName, subkeyPrefix) {
				return true
			}
		}

		return false
	} else {
		// The key we were looking for has been found
		return true
	}
}

func checkRegistry() (bool, string) {

	hyperVKeys := []string{
		`HKLM\SOFTWARE\Microsoft\Hyper-V`,
		`HKLM\SOFTWARE\Microsoft\VirtualMachine`,
		`HKLM\SOFTWARE\Microsoft\Virtual Machine\Guest\Parameters`,
		`HKLM\SYSTEM\ControlSet001\Services\vmicheartbeat`,
		`HKLM\SYSTEM\ControlSet001\Services\vmicvss`,
		`HKLM\SYSTEM\ControlSet001\Services\vmicshutdown`,
		`HKLM\SYSTEM\ControlSet001\Services\vmicexchange`,
	}

	parallelsKeys := []string{
		`HKLM\SYSTEM\CurrentControlSet\Enum\PCI\VEN_1AB8*`,
	}

	virtualBoxKeys := []string{
		`HKLM\SYSTEM\CurrentControlSet\Enum\PCI\VEN_80EE*`,
		`HKLM\HARDWARE\ACPI\DSDT\VBOX__`,
		`HKLM\HARDWARE\ACPI\FADT\VBOX__`,
		`HKLM\HARDWARE\ACPI\RSDT\VBOX__`,
		`HKLM\SOFTWARE\Oracle\VirtualBox Guest Additions`,
		`HKLM\SYSTEM\ControlSet001\Services\VBoxGuest`,
		`HKLM\SYSTEM\ControlSet001\Services\VBoxMouse`,
		`HKLM\SYSTEM\ControlSet001\Services\VBoxService`,
		`HKLM\SYSTEM\ControlSet001\Services\VBoxSF`,
		`HKLM\SYSTEM\ControlSet001\Services\VBoxVideo`,
	}

	virtualPCKeys := []string{
		`HKLM\SYSTEM\CurrentControlSet\Enum\PCI\VEN_5333*`,
		`HKLM\SYSTEM\ControlSet001\Services\vpcbus`,
		`HKLM\SYSTEM\ControlSet001\Services\vpc-s3`,
		`HKLM\SYSTEM\ControlSet001\Services\vpcuhub`,
		`HKLM\SYSTEM\ControlSet001\Services\msvmmouf`,
	}

	vmwareKeys := []string{
		`HKLM\SYSTEM\CurrentControlSet\Enum\PCI\VEN_15AD*`,
		`HKCU\SOFTWARE\VMware, Inc.\VMware Tools`,
		`HKLM\SOFTWARE\VMware, Inc.\VMware Tools`,
		`HKLM\SYSTEM\ControlSet001\Services\vmdebug`,
		`HKLM\SYSTEM\ControlSet001\Services\vmmouse`,
		`HKLM\SYSTEM\ControlSet001\Services\VMTools`,
		`HKLM\SYSTEM\ControlSet001\Services\VMMEMCTL`,
		`HKLM\SYSTEM\ControlSet001\Services\vmware`,
		`HKLM\SYSTEM\ControlSet001\Services\vmci`,
		`HKLM\SYSTEM\ControlSet001\Services\vmx86`,
		`HKLM\SYSTEM\CurrentControlSet\Enum\IDE\CdRomNECVMWar_VMware_IDE_CD*`,
		`HKLM\SYSTEM\CurrentControlSet\Enum\IDE\CdRomNECVMWar_VMware_SATA_CD*`,
		`HKLM\SYSTEM\CurrentControlSet\Enum\IDE\DiskVMware_Virtual_IDE_Hard_Drive*`,
		`HKLM\SYSTEM\CurrentControlSet\Enum\IDE\DiskVMware_Virtual_SATA_Hard_Drive*`,
	}

	xenKeys := []string{
		`HKLM\HARDWARE\ACPI\DSDT\xen`,
		`HKLM\HARDWARE\ACPI\FADT\xen`,
		`HKLM\HARDWARE\ACPI\RSDT\xen`,
		`HKLM\SYSTEM\ControlSet001\Services\xenevtchn`,
		`HKLM\SYSTEM\ControlSet001\Services\xennet`,
		`HKLM\SYSTEM\ControlSet001\Services\xennet6`,
		`HKLM\SYSTEM\ControlSet001\Services\xensvc`,
		`HKLM\SYSTEM\ControlSet001\Services\xenvdb`,
	}

	allKeys := [][]string{hyperVKeys, parallelsKeys, virtualBoxKeys, virtualPCKeys, vmwareKeys, xenKeys}

	for _, keys := range allKeys {
		for _, key := range keys {
			if doesRegistryKeyExist(key) {
				return true, key
			}
		}
	}

	return false, "none"
}

/*
	Public function returning true if a VM is detected.
	If so, a non-empty string is also returned to tell how it was detected.
*/
func IsRunningInVirtualMachine() (bool, string) {
	/*if vmDetected, how := CommonChecks(); vmDetected {
		return vmDetected, how
	}*/

	if vmDetected, registryKey := checkRegistry(); vmDetected {
		return vmDetected, fmt.Sprintf("Registry key (%v)", registryKey)
	}

	return false, "nothing"
}
