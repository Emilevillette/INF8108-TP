package main

import (
	"fmt"
	"golang.org/x/sys/windows"
	"strings"
	"syscall"
	"unsafe"
)

var (
	wlanapi                = windows.NewLazySystemDLL("wlanapi.dll")
	procWlanOpenHandle     = wlanapi.NewProc("WlanOpenHandle")
	procWlanEnumInterfaces = wlanapi.NewProc("WlanEnumInterfaces")
	procWlanGetProfileList = wlanapi.NewProc("WlanGetProfileList")
	procWlanGetProfile     = wlanapi.NewProc("WlanGetProfile")
	procWlanFreeMemory     = wlanapi.NewProc("WlanFreeMemory")
	procWlanCloseHandle    = wlanapi.NewProc("WlanCloseHandle")
)

type WLAN_INTERFACE_INFO_LIST struct {
	dwNumberOfItems uint32
	dwIndex         uint32
	InterfaceInfo   [1]WLAN_INTERFACE_INFO
}

type WLAN_INTERFACE_INFO struct {
	InterfaceGuid           windows.GUID
	strInterfaceDescription [256]uint16
	isState                 uint32
}

type WLAN_PROFILE_INFO_LIST struct {
	dwNumberOfItems uint32
	dwIndex         uint32
	ProfileInfo     [1]WLAN_PROFILE_INFO
}

type WLAN_PROFILE_INFO struct {
	strProfileName [256]uint16
	dwFlags        uint32
}

func wifi_main() ([]string, error) {
	var ssidPasswordPairs []string
	handle, err := wlanOpenHandle()
	if err != nil {
		fmt.Printf("Error opening WLAN handle: %v\n", err)
		return nil, err
	}
	defer wlanCloseHandle(handle)

	interfaces, err := wlanEnumInterfaces(handle)
	if err != nil {
		fmt.Printf("Error enumerating WLAN interfaces: %v\n", err)
		return nil, err
	}

	for i := uint32(0); i < interfaces.dwNumberOfItems; i++ {
		iface := interfaces.InterfaceInfo[i]
		profileList, err := wlanGetProfileList(handle, &iface.InterfaceGuid)
		if err != nil {
			fmt.Printf("Error getting profile list: %v\n", err)
			continue
		}

		profileInfoPtr := uintptr(unsafe.Pointer(&profileList.ProfileInfo[0]))
		for j := uint32(0); j < profileList.dwNumberOfItems; j++ {
			profileInfo := (*WLAN_PROFILE_INFO)(unsafe.Pointer(profileInfoPtr))
			profileName := syscall.UTF16ToString(profileInfo.strProfileName[:])

			profileXML, err := wlanGetProfile(handle, &iface.InterfaceGuid, profileName)
			if err != nil {
				fmt.Printf("Error getting profile %s: %v\n", profileName, err)
				profileInfoPtr += unsafe.Sizeof(WLAN_PROFILE_INFO{})
				continue
			}

			if profileName == "eduroam" {
				profileInfoPtr += unsafe.Sizeof(WLAN_PROFILE_INFO{})
				continue
			}
			password, err := extractProfileInfo(profileXML)
			if err != nil {
				fmt.Printf("Error extracting profile info: %v\n", err)
				profileInfoPtr += unsafe.Sizeof(WLAN_PROFILE_INFO{})
				continue
			}
			ssidPasswordPairs = append(ssidPasswordPairs, fmt.Sprintf("%s:%s", profileName, password))

			profileInfoPtr += unsafe.Sizeof(WLAN_PROFILE_INFO{})
		}

		wlanFreeMemory(unsafe.Pointer(profileList))
	}
	return ssidPasswordPairs, nil
}

func extractProfileInfo(profileXML string) (string, error) {
	authType := extractBetween(profileXML, "<authentication>", "</authentication>")
	if strings.Contains(authType, "WPA2") && strings.Contains(profileXML, "<EAPConfig>") {
		return "", fmt.Errorf("WPA2-Enterprise network")
	} else {
		keyMaterial := extractBetween(profileXML, "<keyMaterial>", "</keyMaterial>")
		if keyMaterial != "" {
			return keyMaterial, nil
		} else if authType == "open" {
			return "", nil
		} else {
			return "", fmt.Errorf("key material not found in profile XML")
		}
	}
}

func extractBetween(s, start, end string) string {
	startIndex := strings.Index(s, start)
	if startIndex == -1 {
		return ""
	}
	startIndex += len(start)
	endIndex := strings.Index(s[startIndex:], end)
	if endIndex == -1 {
		return ""
	}
	return s[startIndex : startIndex+endIndex]
}

func wlanOpenHandle() (handle windows.Handle, err error) {
	var negotiatedVersion uint32
	r0, _, _ := procWlanOpenHandle.Call(
		uintptr(2), // Client version
		0,
		uintptr(unsafe.Pointer(&negotiatedVersion)),
		uintptr(unsafe.Pointer(&handle)),
	)
	if r0 != 0 {
		return 0, fmt.Errorf("WlanOpenHandle failed with error: %d", r0)
	}
	return handle, nil
}

func wlanEnumInterfaces(handle windows.Handle) (*WLAN_INTERFACE_INFO_LIST, error) {
	var interfaceList *WLAN_INTERFACE_INFO_LIST
	r0, _, _ := procWlanEnumInterfaces.Call(
		uintptr(handle),
		0,
		uintptr(unsafe.Pointer(&interfaceList)),
	)
	if r0 != 0 {
		return nil, fmt.Errorf("WlanEnumInterfaces failed with error: %d", r0)
	}
	return interfaceList, nil
}

func wlanGetProfileList(handle windows.Handle, interfaceGuid *windows.GUID) (*WLAN_PROFILE_INFO_LIST, error) {
	var profileList *WLAN_PROFILE_INFO_LIST
	r0, _, _ := procWlanGetProfileList.Call(
		uintptr(handle),
		uintptr(unsafe.Pointer(interfaceGuid)),
		0,
		uintptr(unsafe.Pointer(&profileList)),
	)
	if r0 != 0 {
		return nil, fmt.Errorf("WlanGetProfileList failed with error: %d", r0)
	}
	return profileList, nil
}

func wlanGetProfile(handle windows.Handle, interfaceGuid *windows.GUID, profileName string) (string, error) {
	var flags uint32 = 4 // WLAN_PROFILE_GET_PLAINTEXT_KEY
	var accessAllowed uint32
	var profileXML *uint16
	profileNamePtr, _ := syscall.UTF16PtrFromString(profileName)

	r0, _, _ := procWlanGetProfile.Call(
		uintptr(handle),
		uintptr(unsafe.Pointer(interfaceGuid)),
		uintptr(unsafe.Pointer(profileNamePtr)),
		0,
		uintptr(unsafe.Pointer(&profileXML)),
		uintptr(unsafe.Pointer(&flags)),
		uintptr(unsafe.Pointer(&accessAllowed)),
	)
	if r0 != 0 {
		return "", fmt.Errorf("WlanGetProfile failed with error: %d", r0)
	}
	defer wlanFreeMemory(unsafe.Pointer(profileXML))

	return windows.UTF16PtrToString(profileXML), nil
}

func wlanFreeMemory(mem unsafe.Pointer) {
	procWlanFreeMemory.Call(uintptr(mem))
}

func wlanCloseHandle(handle windows.Handle) {
	procWlanCloseHandle.Call(uintptr(handle), 0)
}
