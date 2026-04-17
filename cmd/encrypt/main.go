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
    plaintext := flag.String("plaintext", "", "plaintext to encrypt")
    decrypt := flag.Bool("decrypt", false, "enable decrypt mode")
    ciphertext := flag.String("ciphertext", "", "ciphertext to decrypt")
    flag.Parse()

    if *key == "" {
        flag.Usage()
        os.Exit(1)
    }

    if *decrypt {
        if *ciphertext == "" {
            fmt.Fprintln(os.Stderr, "error: -ciphertext is required in decrypt mode")
            flag.Usage()
            os.Exit(1)
        }
        result, err := crypto.Decrypt(*key, *ciphertext)
        if err != nil {
            fmt.Fprintf(os.Stderr, "decrypt failed: %v\n", err)
            os.Exit(1)
        }
        fmt.Print(result)
    } else {
        if *plaintext == "" {
            fmt.Fprintln(os.Stderr, "error: -plaintext is required in encrypt mode")
            flag.Usage()
            os.Exit(1)
        }
        result, err := crypto.Encrypt(*key, *plaintext)
        if err != nil {
            fmt.Fprintf(os.Stderr, "encrypt failed: %v\n", err)
            os.Exit(1)
        }
        fmt.Print(result)
    }
}
