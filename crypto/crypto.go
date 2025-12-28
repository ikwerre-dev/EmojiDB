package crypto

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"
	"sort"
	"strings"
)

var (
	EmojiAlphabet = []string{
		"😀", "😁", "😂", "🤣", "😃", "😄", "😅", "😆", "😉", "😊", "😋", "😎", "😍", "😘", "🥰", "😗",
		"😙", "😚", "☺️", "🙂", "🤗", "🤩", "🤔", "🤨", "😐", "😑", "😶", "🙄", "😏", "😣", "😥", "😮",
		"🤐", "😯", "😪", "😫", "😴", "😌", "😛", "😜", "😝", "🤤", "😒", "😓", "😔", "😕", "🙃", "🤑",
		"😲", "☹️", "🙁", "😖", "😞", "😟", "😤", "😢", "😭", "😦", "😧", "😨", "😩", "🤯", "😬", "😰",
		"😱", "🥵", "🥶", "😳", "🤪", "😵", "😡", "😠", "🤬", "😷", "🤒", "🤕", "🤢", "🤮", "🤧", "😇",
		"🤠", "🤡", "🥳", "🥴", "🥺", "🤥", "🤫", "🤭", "🧐", "🤓", "😈", "👿", "👹", "👺", "💀", "👻",
		"👽", "🤖", "💩", "😺", "😸", "😹", "😻", "😼", "😽", "🙀", "😿", "😾", "🙈", "🙉", "🙊", "💋",
		"💌", "💘", "💝", "💖", "💗", "💓", "💞", "💕", "💟", "❣️", "💔", "❤️", "🧡", "💛", "💚", "💙",
		"💜", "🤎", "🖤", "🤍", "💯", "💢", "💥", "💫", "💦", "💨", "🕳️", "💣", "💬", "👁️‍🗨️", "🗨️", "🗯️",
		"💭", "💤", "👋", "🤚", "🖐️", "✋", "🖖", "👌", "🤏", "✌️", "🤞", "🤟", "🤘", "🤙", "👈", "👉",
		"👆", "🖕", "👇", "☝️", "👍", "👎", "✊", "👊", "🤛", "🤜", "👏", "🙌", "👐", "🤲", "🤝", "🙏",
		"✍️", "💅", "🤳", "💪", "🦾", "🦵", "🦿", "👣", "👂", "🦻", "👃", "🧠", "🦷", "🦴", "👀", "👁️",
		"👅", "👄", "👶", "🧒", "👦", "👧", "🧑", "👱", "👨", "🧔", "👩", "🧓", "👴", "👵", "👨‍⚕️", "👩‍⚕️",
		"👨‍🎓", "👩‍🎓", "👨‍🏫", "👩‍🏫", "👨‍⚖️", "👩‍⚖️", "👨‍🌾", "👩‍🌾", "👨‍🍳", "👩‍🍳", "👨‍🔧", "👩‍🔧", "👨‍🏭", "👩‍🏭", "👨‍💼", "👩‍💼",
		"👨‍🔬", "👩‍🔬", "👨‍💻", "👩‍💻", "👨‍🎤", "👩‍🎤", "👨‍🎨", "👩‍🎨", "👨‍✈️", "👩‍✈️", "👨‍🚀", "👩‍🚀", "👨‍🚒", "👩‍🚒", "👮", "🕵️",
		"💂", "👷", "🤴", "👸", "👳", "👲", "🧕", "🤵", "👰", "🤰", "🤱", "👼", "🎅", "🤶", "🦸", "🦹",
	}
	cachedSorted []eData
	encodingMap  [256]string
)

func init() {
	for i, e := range EmojiAlphabet {
		if i < 256 {
			cachedSorted = append(cachedSorted, eData{e, byte(i)})
			encodingMap[i] = e
		}
	}
	sort.Slice(cachedSorted, func(i, j int) bool {
		return len(cachedSorted[i].s) > len(cachedSorted[j].s)
	})
}

func DeriveKey(key string) []byte {
	hash := sha256.Sum256([]byte(key))
	return hash[:]
}

func Encrypt(data []byte, key string) ([]byte, error) {
	block, err := aes.NewCipher(DeriveKey(key))
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, data, nil), nil
}

func Decrypt(ciphertext []byte, key string) ([]byte, error) {
	block, err := aes.NewCipher(DeriveKey(key))
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func EncodeToEmojis(data []byte) string {
	var sb strings.Builder
	for _, b := range data {
		sb.WriteString(encodingMap[b])
	}
	return sb.String()
}

func DecodeFromEmojis(s string) ([]byte, error) {
	sorted := getSortedAlphabet()
	var result []byte
	remaining := s
	for len(remaining) > 0 {
		found := false
		for _, ed := range sorted {
			if strings.HasPrefix(remaining, ed.s) {
				result = append(result, ed.base)
				remaining = remaining[len(ed.s):]
				found = true
				break
			}
		}
		if !found {
			return nil, errors.New("invalid emoji in payload")
		}
	}
	return result, nil
}

type eData struct {
	s    string
	base byte
}

func getSortedAlphabet() []eData {
	return cachedSorted
}

// DecodeOne reads one emoji from the reader and returns the original byte
func DecodeOne(r *bufio.Reader) (byte, error) {
	sorted := getSortedAlphabet()
	var buf []byte
	var lastMatch *eData
	var lastMatchLen int

	for {
		b, err := r.ReadByte()
		if err != nil {
			if err == io.EOF && lastMatch != nil {
				return lastMatch.base, nil
			}
			return 0, err
		}
		buf = append(buf, b)
		s := string(buf)

		foundPrefix := false
		var matchFound bool
		var matchVal byte
		for _, ed := range sorted {
			if s == ed.s {
				matchFound = true
				matchVal = ed.base
			}
			if strings.HasPrefix(ed.s, s) {
				foundPrefix = true
			}
		}

		if matchFound {
			lastMatch = &eData{s: s, base: matchVal}
			lastMatchLen = len(buf)
		}

		if !foundPrefix {
			if lastMatch != nil {
				extra := len(buf) - lastMatchLen
				for i := 0; i < extra; i++ {
					r.UnreadByte()
				}
				return lastMatch.base, nil
			}
			return 0, errors.New("invalid emoji sequence")
		}

		if len(buf) > 64 {
			return 0, errors.New("sequence too long")
		}
	}
}
