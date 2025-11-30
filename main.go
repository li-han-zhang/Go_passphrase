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
    genBinary := flag.Bool("b", false, "Generate binary.txt only")
    useBinary := flag.Bool("p", false, "Generate passphrase from binary.txt")
    showHelp := flag.Bool("h", false, "Show help message")
    showQRCode := flag.Bool("q", false, "Generate QR code of passphrase from binary.txt")
    inspectWord := flag.String("i", "", "Inspect a word or 11-bit binary")

    flag.Parse()

    if !*genBinary && !*useBinary && !*showQRCode && !*showHelp && *inspectWord == "" {
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

    // -i WORD / -i BINARY
    if *inspectWord != "" {
        showWordInfo(*inspectWord, wordList)
        return
    }

    // -b → generate binary
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

    // -p → passphrase
    if *useBinary {
        passphrase := generatePassphraseFromBinary(wordList)
        fmt.Println("Passphrase:")
        fmt.Println(passphrase)
    }

    // -q → QR Code
    if *showQRCode {
        passphrase := generatePassphraseFromBinary(wordList)
        fmt.Println("Passphrase QR Code:")
        qr, err := qrcode.New(passphrase, qrcode.Low)
        if err != nil {
            log.Fatalf("Error generating QR code: %v", err)
        }
        fmt.Println(qr.ToSmallString(false))
    }
}

//
// -------------------------
//   -i 处理逻辑（双模式）
// -------------------------
//

func showWordInfo(input string, wordList []string) {
    input = strings.TrimSpace(strings.ToLower(input))

    // 1. 输入是二进制？
    if isBinary(input) {
        // 自动补齐 11 位
        if len(input) < 11 {
            input = fmt.Sprintf("%011s", input)
        }
        if len(input) != 11 {
            fmt.Println("Error: binary must be <= 11 bits.")
            return
        }

        idx := binaryToInt(input)
        if idx < 0 || idx >= 2048 {
            fmt.Println("Error: binary out of range (0–2047).")
            return
        }

        fmt.Println("Binary:", input)
        fmt.Println("Index:", idx)
        fmt.Println("Word:", wordList[idx])
        return
    }

    // 2. 否则输入是单词
    for i, w := range wordList {
        if w == input {
            fmt.Println("Word:", input)
            fmt.Println("Index:", i)
            fmt.Printf("Binary: %011b\n", i)
            return
        }
    }

    fmt.Printf("Error: '%s' is neither a valid word nor a valid binary.\n", input)
}

func isBinary(s string) bool {
    for _, c := range s {
        if c != '0' && c != '1' {
            return false
        }
    }
    return len(s) > 0
}

func binaryToInt(bin string) int {
    n := 0
    for _, c := range bin {
        n <<= 1
        if c == '1' {
            n |= 1
        }
    }
    return n
}

//
// -------------------------
//      Helper Functions
// -------------------------
//

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
    fmt.Println("  -i WORD   Show WORD's index and 11-bit binary")
    fmt.Println("  -i BIN    Show BIN's index and corresponding word")
    fmt.Println("  -h        Show this help message")
}

func loadWordList() []string {
    lines := strings.Split(wordListText, "\n")
    words := make([]string, 0, 2048)
    for _, w := range lines {
        w = strings.TrimSpace(w)
        w = strings.TrimPrefix(w, "\ufeff")
        if w != "" {
            words = append(words, w)
        }
    }
    return words
}

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

func generatePassphraseFromBinary(wordList []string) string {
    if _, err := os.Stat("binary.txt"); os.IsNotExist(err) {
        log.Fatalf("Error: binary.txt not found. Use -b first.")
    }
    bits, err := readBinaryFile("binary.txt")
    if err != nil {
        log.Fatalf("Error reading binary.txt: %v", err)
    }
    csBits := checksumBits(bits)
    allBits := append(bits, csBits...)
    return generateMnemonic(allBits, wordList)
}

func generateMnemonic(bits []bool, wordList []string) string {
    wordCount := len(bits) / 11
    words := make([]string, 0, wordCount)
    for i := 0; i < wordCount; i++ {
        index := bitsToInt(bits[i*11 : (i+1)*11])
        words = append(words, wordList[index])
    }
    return strings.Join(words, " ")
}

func checksumBits(entropyBits []bool) []bool {
    entropy := bitsToBytes(entropyBits)
    hash := sha256.Sum256(entropy)
    csLen := len(entropyBits) / 32
    return bytesToBits(hash[:])[:csLen]
}

func bytesToBits(b []byte) []bool {
    bits := make([]bool, 0, len(b)*8)
    for _, by := range b {
        for i := 7; i >= 0; i-- {
            bits = append(bits, ((by>>i)&1) == 1)
        }
    }
    return bits
}

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
