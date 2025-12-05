package main

import (
	"fmt"
	"log"

	"ironnode/pkg/crypto"
)

func main() {
	fmt.Println("=== Generating Encryption Key for Wallet Private Keys ===")
	fmt.Println()

	key, err := crypto.GenerateEncryptionKey()
	if err != nil {
		log.Fatal("Failed to generate encryption key:", err)
	}

	fmt.Println("✅ Encryption key generated successfully!")
	fmt.Println()
	fmt.Println("Add this line to your .env file:")
	fmt.Println()
	fmt.Printf("ENCRYPTION_KEY=%s\n", key)
	fmt.Println()
	fmt.Println("⚠️  IMPORTANT: Keep this key secure! Without it, you won't be able to decrypt private keys.")
	fmt.Println("⚠️  Never commit this key to version control!")
}
