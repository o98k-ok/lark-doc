package auth

/*
#cgo LDFLAGS: -framework Security -framework CoreFoundation
#include <Security/Security.h>
#include <stdlib.h>
#include <string.h>

static CFMutableDictionaryRef query(const char *account) {
    CFMutableDictionaryRef q = CFDictionaryCreateMutable(NULL, 0, &kCFTypeDictionaryKeyCallBacks, &kCFTypeDictionaryValueCallBacks);
    CFStringRef a = CFStringCreateWithCString(NULL, account, kCFStringEncodingUTF8);
    CFDictionarySetValue(q, kSecClass, kSecClassGenericPassword);
    CFDictionarySetValue(q, kSecAttrService, CFSTR("com.o98k.lark-doc"));
    CFDictionarySetValue(q, kSecAttrAccount, a);
    CFRelease(a);
    return q;
}
static int readSecret(const char *account, void **bytes, long *length) {
    CFMutableDictionaryRef q = query(account);
    CFDictionarySetValue(q, kSecReturnData, kCFBooleanTrue);
    CFTypeRef value = NULL;
    OSStatus status = SecItemCopyMatching(q, &value);
    CFRelease(q);
    if (status != errSecSuccess) return status;
    *length = CFDataGetLength((CFDataRef)value);
    *bytes = malloc(*length);
    if (!*bytes) { CFRelease(value); return errSecAllocate; }
    memcpy(*bytes, CFDataGetBytePtr((CFDataRef)value), *length);
    CFRelease(value);
    return 0;
}
static int writeSecret(const char *account, const char *secret) {
    CFMutableDictionaryRef q = query(account);
    CFDataRef value = CFDataCreate(NULL, (const UInt8 *)secret, strlen(secret));
    CFMutableDictionaryRef attrs = CFDictionaryCreateMutable(NULL, 0, &kCFTypeDictionaryKeyCallBacks, &kCFTypeDictionaryValueCallBacks);
    CFDictionarySetValue(attrs, kSecValueData, value);
    OSStatus status = SecItemUpdate(q, attrs);
    if (status == errSecItemNotFound) {
        CFDictionarySetValue(q, kSecValueData, value);
        status = SecItemAdd(q, NULL);
    }
    CFRelease(attrs); CFRelease(value); CFRelease(q);
    return status;
}
static int deleteSecret(const char *account) {
    CFMutableDictionaryRef q = query(account);
    OSStatus status = SecItemDelete(q);
    CFRelease(q);
    return status == errSecItemNotFound ? 0 : status;
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

const useKeychain = true

func keychainLoad(path string) (string, error) {
	account := C.CString(path)
	defer C.free(unsafe.Pointer(account))
	var bytes unsafe.Pointer
	var length C.long
	status := C.readSecret(account, &bytes, &length)
	if status != 0 {
		return "", fmt.Errorf("Keychain unavailable (%d); unlock Keychain or run lark-doc login", status)
	}
	defer C.free(bytes)
	return C.GoStringN((*C.char)(bytes), C.int(length)), nil
}

func keychainSave(path, session string) error {
	account, secret := C.CString(path), C.CString(session)
	defer C.free(unsafe.Pointer(account))
	defer C.free(unsafe.Pointer(secret))
	if status := C.writeSecret(account, secret); status != 0 {
		return fmt.Errorf("Keychain save failed (%d)", status)
	}
	return nil
}

func keychainDelete(path string) error {
	account := C.CString(path)
	defer C.free(unsafe.Pointer(account))
	if status := C.deleteSecret(account); status != 0 {
		return fmt.Errorf("Keychain delete failed (%d)", status)
	}
	return nil
}
