// CLI encryption tool: AES-256-GCM encrypt/decrypt operations
package main

import (
    "flag"
    "fmt"
    "os"

    "github.com/vincent119/audit-notifier/internal/crypto"
)

func main() {
    key := flag.String("key", "", "encryption key (required)")
    encrypt := flag.String("encrypt", "", "plaintext to encrypt")
    decrypt := flag.String("decrypt", "", "ciphertext to decrypt")
    flag.Parse()

    if *key == "" {
        flag.Usage()
        os.Exit(1)
    }

    switch {
    case *encrypt != "":
        result, err := crypto.Encrypt(*key, *encrypt)
        if err != nil {
            fmt.Fprintf(os.Stderr, "encrypt failed: %v\n", err)
            os.Exit(1)
        }
        fmt.Print(result)
    case *decrypt != "":
        result, err := crypto.Decrypt(*key, *decrypt)
        if err != nil {
            fmt.Fprintf(os.Stderr, "decrypt failed: %v\n", err)
            os.Exit(1)
        }
        fmt.Print(result)
    default:
        fmt.Fprintln(os.Stderr, "error: -encrypt or -decrypt is required")
        flag.Usage()
        os.Exit(1)
    }
}
