package crypto

import (
    "testing"
)

// TestEncryptDecrypt 驗證加密後解密能還原明文。
func TestEncryptDecrypt(t *testing.T) {
    key := "test-secret-key"
    plaintext := "Hello, World!"

    encrypted, err := Encrypt(key, plaintext)
    if err != nil {
        t.Fatalf("加密失敗: %v", err)
    }

    decrypted, err := Decrypt(key, encrypted)
    if err != nil {
        t.Fatalf("解密失敗: %v", err)
    }

    if decrypted != plaintext {
        t.Errorf("解密結果不一致: got %q, want %q", decrypted, plaintext)
    }
}

// TestDecryptWrongKey 驗證使用錯誤 key 解密會失敗。
func TestDecryptWrongKey(t *testing.T) {
    encrypted, err := Encrypt("correct-key", "secret data")
    if err != nil {
        t.Fatalf("加密失敗: %v", err)
    }

    _, err = Decrypt("wrong-key", encrypted)
    if err == nil {
        t.Error("使用錯誤 key 解密應回傳 error")
    }
}

// TestEncryptEmptyPlaintext 驗證空明文回傳 error。
func TestEncryptEmptyPlaintext(t *testing.T) {
    _, err := Encrypt("key", "")
    if err == nil {
        t.Error("空明文應回傳 error")
    }
}

// TestDecryptEmptyCiphertext 驗證空密文回傳 error。
func TestDecryptEmptyCiphertext(t *testing.T) {
    _, err := Decrypt("key", "")
    if err == nil {
        t.Error("空密文應回傳 error")
    }
}

// TestEncryptRandomSalt 驗證同一明文加密兩次結果不同（隨機 salt）。
func TestEncryptRandomSalt(t *testing.T) {
    key := "test-key"
    plaintext := "same plaintext"

    enc1, err := Encrypt(key, plaintext)
    if err != nil {
        t.Fatalf("第一次加密失敗: %v", err)
    }

    enc2, err := Encrypt(key, plaintext)
    if err != nil {
        t.Fatalf("第二次加密失敗: %v", err)
    }

    if enc1 == enc2 {
        t.Error("同一明文加密兩次結果應不同")
    }
}
