package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"
)

type Book struct {
	Code        string `json:"code"`
	Title       string `json:"title"`
	Author      string `json:"author"`
	Publisher   string `json:"publisher"`
	Description string `json:"description"`
}

var bookList []Book

func main() {
	loadBooksFromJSON()

	for {
		fmt.Println("=== Menu Utama ===")
		fmt.Println("1. Tambah Buku")
		fmt.Println("2. Tampilkan List Buku")
		fmt.Println("3. Hapus Buku")
		fmt.Println("4. Edit Buku")
		fmt.Println("5. Print Buku")
		fmt.Println("6. Keluar")
		fmt.Print("Pilih menu: ")

		var choice int
		fmt.Scanln(&choice)

		switch choice {
		case 1:
			addBook()
		case 2:
			listBooks()
		case 3:
			deleteBook()
		case 4:
			editBook()
		case 5:
			printBook()
			continueLoop()
		case 6:
			fmt.Println("Terima kasih telah menggunakan aplikasi ini.")
			return
		default:
			fmt.Println("Menu tidak valid. Silakan pilih lagi.")
		}
	}
}

func loadBooksFromJSON() {
	files, err := ioutil.ReadDir("books")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for _, file := range files {
		if !file.IsDir() {
			fileName := file.Name()
			if strings.HasPrefix(fileName, "book-") && strings.HasSuffix(fileName, ".json") {
				data, err := ioutil.ReadFile("books/" + fileName)
				if err != nil {
					fmt.Println("Error:", err)
					continue
				}
				var book Book
				err = json.Unmarshal(data, &book)
				if err != nil {
					fmt.Println("Error:", err)
					continue
				}
				bookList = append(bookList, book)
			}
		}
	}
}

func addBook() {
	for {
		var newBook Book

		fmt.Println("=== Tambah Buku ===")
		fmt.Print("Judul: ")
		newBook.Title = getInput()
		fmt.Print("Penulis: ")
		newBook.Author = getInput()
		fmt.Print("Penerbit: ")
		newBook.Publisher = getInput()
		fmt.Print("Deskripsi: ")
		newBook.Description = getInput()

		for {
			newBook.Code = generateUniqueCode()
			if !isBookCodeUsed(newBook.Code) {
				break
			}
		}

		bookList = append(bookList, newBook)

		bookJSON, _ := json.MarshalIndent(newBook, "", "    ")
		filename := fmt.Sprintf("books/book-%s.json", newBook.Code)
		err := ioutil.WriteFile(filename, bookJSON, 0644)
		if err != nil {
			fmt.Println("Error:", err)
		}

		fmt.Println("Buku berhasil ditambahkan!")

		fmt.Println("Apakah Anda ingin menambahkan buku lain? (y/n):")
		var input string
		fmt.Scanln(&input)
		if strings.ToLower(input) != "y" {
			break
		}
	}
}

func getInput() string {
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func generateUniqueCode() string {
	return strings.Replace(uuid.New().String(), "-", "", -1)[:6]
}

func isBookCodeUsed(code string) bool {
	for _, book := range bookList {
		if book.Code == code {
			fmt.Println("Kode buku sudah digunakan. Silakan coba lagi.")
			return true
		}
	}
	return false
}

func listBooks() {
	fmt.Println("=== List Buku ===")
	if len(bookList) == 0 {
		fmt.Println("Tidak ada buku yang tersedia.")
		return
	}

	for _, book := range bookList {
		fmt.Printf("Kode: %s | Judul: %s | Penulis: %s | Penerbit: %s\n", book.Code, book.Title, book.Author, book.Publisher)
	}
}

func deleteBook() {
	fmt.Println("=== Hapus Buku ===")
	fmt.Print("Masukkan kode buku yang ingin dihapus: ")
	code := getInput()

	for i, book := range bookList {
		if book.Code == code {
			filename := fmt.Sprintf("books/book-%s.json", code)
			err := os.Remove(filename)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
			bookList = append(bookList[:i], bookList[i+1:]...)
			fmt.Println("Buku berhasil dihapus.")
			return
		}
	}

	fmt.Println("Buku tidak ditemukan.")
}

func editBook() {
	fmt.Println("=== Edit Buku ===")
	fmt.Print("Masukkan kode buku yang ingin diubah: ")
	code := getInput()

	for i, book := range bookList {
		if book.Code == code {
			fmt.Println("Masukkan informasi baru untuk buku ini:")
			fmt.Print("Judul: ")
			bookList[i].Title = getInput()
			fmt.Print("Penulis: ")
			bookList[i].Author = getInput()
			fmt.Print("Penerbit: ")
			bookList[i].Publisher = getInput()
			fmt.Print("Deskripsi: ")
			bookList[i].Description = getInput()

			bookJSON, _ := json.MarshalIndent(bookList[i], "", "    ")
			filename := fmt.Sprintf("books/book-%s.json", code)
			err := ioutil.WriteFile(filename, bookJSON, 0644)
			if err != nil {
				fmt.Println("Error:", err)
			}
			fmt.Println("Buku berhasil diubah.")
			return
		}
	}

	fmt.Println("Buku tidak ditemukan.")
}

func printBook() {
	fmt.Println("=== Print Buku ===")
	fmt.Println("Pilih buku yang ingin di print (ketik 'all' untuk print semua buku):")
	listBooks()

	var selectedCode string
	fmt.Print("Masukkan kode buku atau 'all': ")
	selectedCode = getInput()

	if selectedCode == "all" {
		ch := make(chan Book)
		wg := sync.WaitGroup{}

		for _, book := range bookList {
			wg.Add(1)
			go func(book Book) {
				defer wg.Done()
				generatePDF(book)
				ch <- book
			}(book)
		}

		go func() {
			wg.Wait()
			close(ch)
		}()

		for range bookList {
			<-ch
		}

		fmt.Println("Semua buku berhasil di print.")
	} else {
		ch := make(chan bool)

		for _, book := range bookList {
			if book.Code == selectedCode {
				go func(book Book) {
					generatePDF(book)
					ch <- true
				}(book)
				break
			}
		}

		<-ch
		fmt.Println("Buku berhasil di print.")
	}
}

func generatePDF(book Book) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "Judul: "+book.Title)
	pdf.Ln(10)
	pdf.Cell(40, 10, "Penulis: "+book.Author)
	pdf.Ln(10)
	pdf.Cell(40, 10, "Penerbit: "+book.Publisher)
	pdf.Ln(10)
	pdf.Cell(40, 10, "Deskripsi: "+book.Description)
	pdf.Ln(10)

	filename := fmt.Sprintf("pdf/book-%s.pdf", book.Code)
	_ = pdf.OutputFileAndClose(filename)
	fmt.Printf("Buku %s berhasil di print.\n", book.Title)
}

func continueLoop() {
	fmt.Println("Kembali ke Menu Utama...")
	time.Sleep(2 * time.Second)
	fmt.Println()
}
