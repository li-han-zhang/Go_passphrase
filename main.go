package main

import (
	"bufio"
	"crypto/rand"
	"crypto/sha256"
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	qrcode "github.com/skip2/go-qrcode"
)

//go:embed embed/english.txt
var wordListText string

func main() {
	// Command line flags
	genBinary := flag.Bool("b", false, "Generate binary.txt only")
	useBinary := flag.Bool("p", false, "Generate passphrase from binary.txt")
	showHelp := flag.Bool("h", false, "Show help message")
	showQRCode := flag.Bool("q", false, "Generate QR code of passphrase from binary.txt")
	flag.Parse()

	// If no flags provided → show help
	if !*genBinary && !*useBinary && !*showHelp && !*showQRCode {
		printHelp()
		return
	}

	if *showHelp {
		printHelp()
		return
	}

	wordList := loadWordList()
	if len(wordList) != 2048 {
		log.Fatalf("Error: word list length %d, expected 2048", len(wordList))
	}

	// -b → generate binary.txt
	if *genBinary {
		entropy := make([]byte, 32)
		_, err := rand.Read(entropy)
		if err != nil {
			log.Fatalf("Error generating entropy: %v", err)
		}
		err = writeBinaryFile("binary.txt", entropy)
		if err != nil {
			log.Fatalf("Error writing binary.txt: %v", err)
		}
		fmt.Println("binary.txt generated successfully.")
	}

	// -p → generate passphrase
	if *useBinary {
		passphrase := generatePassphraseFromBinary(wordList)
		fmt.Println("Passphrase:")
		fmt.Println(passphrase)
	}

	// -q → generate QR code (compact)
	if *showQRCode {
		passphrase := generatePassphraseFromBinary(wordList)
		fmt.Println("Passphrase QR Code:")
		qr, err := qrcode.New(passphrase, qrcode.Low) // Low error correction → smaller
		if err != nil {
			log.Fatalf("Error generating QR code: %v", err)
		}
		fmt.Println(qr.ToSmallString(false)) // compact terminal display
	}
}

// ----------------- Helper Functions -----------------

func printHelp() {
	fmt.Println("passphrase_bitcoin - A 256-bit entropy & BIP39 passphrase generator")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  passphrase_bitcoin [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -b        Generate binary.txt only")
	fmt.Println("  -p        Generate passphrase from binary.txt")
	fmt.Println("  -q        Generate QR code of passphrase from binary.txt")
	fmt.Println("  -h        Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  passphrase_bitcoin -b")
	fmt.Println("      Generate binary.txt using 256-bit system entropy")
	fmt.Println()
	fmt.Println("  passphrase_bitcoin -p")
	fmt.Println("      Generate a BIP39 passphrase from an existing binary.txt")
	fmt.Println()
	fmt.Println("  passphrase_bitcoin -q")
	fmt.Println("      Display passphrase as a compact terminal QR code")
	fmt.Println()
	fmt.Println("  passphrase_bitcoin -b -q")
	fmt.Println("      Generate binary.txt and display the QR code of the passphrase")
}

// loadWordList loads embedded english.txt
func loadWordList() []string {
	lines := strings.Split(wordListText, "\n")
	words := make([]string, 0, 2048)
	for _, w := range lines {
		w = strings.TrimSpace(w)
		w = strings.TrimPrefix(w, "\ufeff") // remove BOM if exists
		if w != "" {
			words = append(words, w)
		}
	}
	return words
}

// writeBinaryFile formats entropy as binary groups (6 per line)
func writeBinaryFile(filename string, entropy []byte) error {
	bits := bytesToBits(entropy)
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := bufio.NewWriter(f)
	groupCount := 0
	for i, b := range bits {
		if b {
			writer.WriteByte('1')
		} else {
			writer.WriteByte('0')
		}

		if (i+1)%11 == 0 {
			writer.WriteByte(' ')
			groupCount++
		}

		if groupCount == 6 {
			writer.WriteByte('\n')
			groupCount = 0
		}
	}
	if groupCount != 0 {
		writer.WriteByte('\n')
	}
	return writer.Flush()
}

// readBinaryFile reads bits from binary.txt
func readBinaryFile(filename string) ([]bool, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var bits []bool
	for scanner.Scan() {
		line := scanner.Text()
		for _, c := range line {
			if c == '0' {
				bits = append(bits, false)
			} else if c == '1' {
				bits = append(bits, true)
			}
		}
	}
	return bits, scanner.Err()
}

// generatePassphraseFromBinary reads binary.txt and generates the passphrase
func generatePassphraseFromBinary(wordList []string) string {
	if _, err := os.Stat("binary.txt"); os.IsNotExist(err) {
		log.Fatalf("Error: binary.txt not found. Please generate it first with -b")
	}
	bits, err := readBinaryFile("binary.txt")
	if err != nil {
		log.Fatalf("Error reading binary.txt: %v", err)
	}
	csBits := checksumBits(bits)
	allBits := append(bits, csBits...)
	return generateMnemonic(allBits, wordList)
}

// generateMnemonic converts bits → 24 BIP39 words
func generateMnemonic(bits []bool, wordList []string) string {
	wordCount := len(bits) / 11
	words := make([]string, 0, wordCount)
	for i := 0; i < wordCount; i++ {
		index := bitsToInt(bits[i*11 : (i+1)*11])
		words = append(words, wordList[index])
	}
	return strings.Join(words, " ")
}

// checksumBits → first (entropyBits / 32) bits of SHA-256(entropy)
func checksumBits(entropyBits []bool) []bool {
	entropy := bitsToBytes(entropyBits)
	hash := sha256.Sum256(entropy)
	csLen := len(entropyBits) / 32
	return bytesToBits(hash[:])[:csLen]
}

// bytes → []bool (MSB first)
func bytesToBits(b []byte) []bool {
	bits := make([]bool, 0, len(b)*8)
	for _, by := range b {
		for i := 7; i >= 0; i-- {
			bits = append(bits, ((by>>i)&1) == 1)
		}
	}
	return bits
}

// []bool → bytes
func bitsToBytes(bits []bool) []byte {
	n := (len(bits) + 7) / 8
	out := make([]byte, n)
	for i, b := range bits {
		if b {
			out[i/8] |= 1 << (7 - uint(i%8))
		}
	}
	return out
}

// []bool → int (11 bits)
func bitsToInt(bits []bool) int {
	n := 0
	for _, b := range bits {
		n <<= 1
		if b {
			n |= 1
		}
	}
	return n
}
