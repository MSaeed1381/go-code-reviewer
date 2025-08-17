package test

import (
	"context"
	"github.com/stretchr/testify/require"
	"go_code_reviewer/services/code-reviewer/internal/parser"
	"os"
	"path/filepath"
	"testing"
)

func TestParser(t *testing.T) {
	t.Run("go code", func(t *testing.T) {
		projectParser := parser.NewProjectParser(map[string]*parser.CodeParser{
			".go": parser.NewCodeParser(parser.LanguageGo),
		})

		dirPath, err := os.MkdirTemp("", "gh-pr-*")
		require.NoError(t, err)
		defer os.RemoveAll(dirPath)

		filePath := filepath.Join(dirPath, "main.go")
		goCode := `package main

import (
	"fmt"
	"strings"
	"time"
)

// --- Snippet 1 --- 
func helloSnippet() {
	fmt.Println("Hello, Go!")
}

// --- Snippet 2 ---
func stringSnippet() {
	str := "golang is fun"
	upper := strings.ToUpper(str)
	fmt.Println("Original:", str)
	fmt.Println("Upper   :", upper)
}

// --- Snippet 3 ---
func loopSnippet() {
	for i := 1; i <= 5; i++ {
		if i%2 == 0 {
			fmt.Println(i, "even")
		} else {
			fmt.Println(i, "odd")
		}
	}
}

// --- Snippet 4 ---
func goroutineSnippet() {
	go fmt.Println("Running in a goroutine!")
	time.Sleep(100 * time.Millisecond)
}

// --- Snippet 5 ---
func mapSnippet() {
	ages := map[string]int{
		"Alice": 23,
		"Bob":   30,
	}
	ages["Charlie"] = 28

	for name, age := range ages {
		fmt.Printf("%s is %d years old\n", name, age)
	}
}

// --- Snippet 6 ---
func main() {
	helloSnippet()
	stringSnippet()
	loopSnippet()
	goroutineSnippet()
	mapSnippet()
}
`
		err = os.WriteFile(filePath, []byte(goCode), 0644)
		require.NoError(t, err)

		snippets, err := projectParser.ParseProject(context.Background(), filePath)
		require.NoError(t, err)
		require.Len(t, snippets, 6)
		require.Equal(t, "go", snippets[0].Language)
		require.Equal(t, "func helloSnippet() {\n\tfmt.Println(\"Hello, Go!\")\n}", snippets[0].Content)
		require.Contains(t, snippets[1].Filename, "main.go")
	})

	t.Run("python code", func(t *testing.T) {
		projectParser := parser.NewProjectParser(map[string]*parser.CodeParser{
			".py": parser.NewCodeParser(parser.LanguagePython),
		})

		dirPath, err := os.MkdirTemp("", "gh-pr-*")
		require.NoError(t, err)
		defer os.RemoveAll(dirPath)

		filePath := filepath.Join(dirPath, "main.py")
		pythonCode := `import threading
import time
from collections import Counter

def hello_snippet():
    print("Hello, Python!")

def list_snippet():
    nums = [1, 2, 3, 4, 5]
    squares = [n**2 for n in nums]
    print("Numbers:", nums)
    print("Squares:", squares)

def loop_snippet():
    for i in range(1, 6):
        if i % 2 == 0:
            print(i, "even")
        else:
            print(i, "odd")

def thread_snippet():
    def worker():
        print("Running in a thread!")
    t = threading.Thread(target=worker)
    t.start()
    t.join()

def counter_snippet():
    words = ["apple", "banana", "apple", "orange", "banana", "apple"]
    counts = Counter(words)
    print("Word counts:", counts)


def main():
    hello_snippet()
    list_snippet()
    loop_snippet()
    thread_snippet()
    counter_snippet()


if __name__ == "__main__":
    main()
`
		err = os.WriteFile(filePath, []byte(pythonCode), 0644)
		require.NoError(t, err)

		snippets, err := projectParser.ParseProject(context.Background(), filePath)
		require.NoError(t, err)
		require.Len(t, snippets, 6)
		require.Equal(t, "python", snippets[0].Language)
		require.Equal(t, "def hello_snippet():\n    print(\"Hello, Python!\")", snippets[0].Content)
		require.Contains(t, snippets[1].Filename, "main.py")
	})
}
