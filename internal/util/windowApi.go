//go:build windows

// 只对window编译
/**
  @author: cilang
  @qq: 1019383856
  @bili: https://space.bilibili.com/433915419
  @gitee: https://gitee.com/OpencvLZG
  @github: https://github.com/OpencvLZG
  @since: 2023/6/17
  @desc: //TODO
**/

package util

import (
	"fmt"
	"syscall"
	"unsafe"
)

const (
	InternetPerConnFlags              = 1
	InternetPerConnProxyServer        = 2
	InternetPerConnProxyBypass        = 3
	InternetOptionRefresh             = 37
	InternetOptionSettingsChanged     = 39
	InternetOptionPerConnectionOption = 75
)

/*
	typedef struct {
	  DWORD dwOption;
	  union {
	    DWORD    dwValue;
	    LPSTR    pszValue;
	    FILETIME ftValue;
	  } Value;
	} INTERNET_PER_CONN_OPTIONA, *LPINTERNET_PER_CONN_OPTIONA;

	typedef struct _FILETIME {
	  DWORD dwLowDateTime;
	  DWORD dwHighDateTime;
	} FILETIME, *PFILETIME, *LPFILETIME;
*/
type InternetPerConnOption struct {
	dwOption uint32
	dwValue  uint64 // 注意 32位 和 64位 struct 和 union 内存对齐
}

type InternetPerConnOptionList struct {
	dwSize        uint32
	pszConnection *uint16
	dwOptionCount uint32
	dwOptionError uint32
	pOptions      uintptr
}

// SetProxy 设置Window代理
// https://blog.csdn.net/leoforbest/article/details/120166881
func SetProxy(proxy string) error {
	winInet, err := syscall.LoadLibrary("Wininet.dll")
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("LoadLibrary Wininet.dll Error: %s", err))
	}
	InternetSetOptionW, err := syscall.GetProcAddress(winInet, "InternetSetOptionW")
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("GetProcAddress InternetQueryOptionW Error: %s", err))
	}

	options := [3]InternetPerConnOption{}
	options[0].dwOption = InternetPerConnFlags
	if proxy == "" {
		options[0].dwValue = 1
	} else {
		options[0].dwValue = 2
	}
	options[1].dwOption = InternetPerConnProxyServer
	options[1].dwValue = uint64(uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(proxy))))
	options[2].dwOption = InternetPerConnProxyBypass
	options[2].dwValue = uint64(uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("localhost;127.*;10.*;172.16.*;172.17.*;172.18.*;172.19.*;172.20.*;172.21.*;172.22.*;172.23.*;172.24.*;172.25.*;172.26.*;172.27.*;172.28.*;172.29.*;172.30.*;172.31.*;172.32.*;192.168.*"))))

	list := InternetPerConnOptionList{}
	list.dwSize = uint32(unsafe.Sizeof(list))
	list.pszConnection = nil
	list.dwOptionCount = 3
	list.dwOptionError = 0
	list.pOptions = uintptr(unsafe.Pointer(&options))

	// https://www.cnpython.com/qa/361707
	callInternetOptionW := func(dwOption uintptr, lpBuffer uintptr, dwBufferLength uintptr) error {
		r1, _, err := syscall.Syscall6(InternetSetOptionW, 4, 0, dwOption, lpBuffer, dwBufferLength, 0, 0)
		if r1 != 1 {
			return err
		}
		return nil
	}

	err = callInternetOptionW(InternetOptionPerConnectionOption, uintptr(unsafe.Pointer(&list)), uintptr(unsafe.Sizeof(list)))
	if err != nil {
		return fmt.Errorf("INTERNET_OPTION_PER_CONNECTION_OPTION Error: %s", err)
	}
	err = callInternetOptionW(InternetOptionSettingsChanged, 0, 0)
	if err != nil {
		return fmt.Errorf("INTERNET_OPTION_SETTINGS_CHANGED Error: %s", err)
	}
	err = callInternetOptionW(InternetOptionRefresh, 0, 0)
	if err != nil {
		return fmt.Errorf("INTERNET_OPTION_REFRESH Error: %s", err)
	}
	return nil
}
