package unittests

import (
	"testing"

	"github.com/libpub/golib/testingutil"
	"github.com/libpub/golib/utils/cryptoes"
)

func TestCryptoTextsAES(t *testing.T) {
	txt := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz01234567890"
	key := "ABCDEFGHIJKLMNOP"
	iv := "0000000000000000"
	encodedBytes, err := cryptoes.AESEncryptECB([]byte(txt), key)
	testingutil.AssertNil(t, err, "cryptoes.AESEncryptECB error")
	decodedBytes, err := cryptoes.AESDecryptECB(encodedBytes, key)
	testingutil.AssertNil(t, err, "cryptoes.AESDecryptECB error")
	testingutil.AssertEquals(t, txt, string(decodedBytes), "cryptoes.AESDecryptECB result")
	encodedBytes, err = cryptoes.AESEncryptCBC([]byte(txt), key, iv)
	testingutil.AssertNil(t, err, "cryptoes.AESEncryptCBC error")
	decodedBytes, err = cryptoes.AESDecryptCBC(encodedBytes, key, iv)
	testingutil.AssertNil(t, err, "cryptoes.AESDecryptCBC error")
	testingutil.AssertEquals(t, txt, string(decodedBytes), "cryptoes.AESDecryptCBC result")
}
