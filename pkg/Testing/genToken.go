package main

import (
	"fmt"
	"os"
)

func main() {
	file, err := os.Create("users.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	for i := 0; i < 3000; i++ {
		line := fmt.Sprintf(
			"user_%d,123456\n",
			i,
		)
		file.WriteString(line)
	}

	fmt.Println("users.txt generated")
}
