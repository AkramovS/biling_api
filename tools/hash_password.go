package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Использование: go run hash_password.go <password>")
		fmt.Println("Пример: go run hash_password.go password123")
		os.Exit(1)
	}

	password := os.Args[1]
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		fmt.Printf("Ошибка генерации хеша: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✅ Хеш пароля успешно сгенерирован (bcrypt cost 12):")
	fmt.Println()
	fmt.Println(string(hash))
	fmt.Println()
	fmt.Println("💡 Используйте этот хеш в SQL запросе:")
	fmt.Printf("INSERT INTO system_accounts (login, password, name) VALUES ('username', '%s', 'Имя');\n", string(hash))
	fmt.Println()
}
