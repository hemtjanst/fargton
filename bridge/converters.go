package bridge

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/lucasb-eyer/go-colorful"
)

// DateTimeToISO8600 formats a time.Time as ISO 8601:2004
func DateTimeToISO8600(t time.Time) string {
	return fmt.Sprintf(
		"%d-%s-%sT%s:%s:%s",
		t.Year(),
		fmt.Sprintf("%02d", t.Month()),
		fmt.Sprintf("%02d", t.Day()),
		fmt.Sprintf("%02d", t.Hour()),
		fmt.Sprintf("%02d", t.Minute()),
		fmt.Sprintf("%02d", t.Second()),
	)
}

// Brightness ensures the value is always between 1-254
func Brightness(bri uint8) uint8 {
	return uint8(math.Min(math.Max(1, float64(bri)), 254))
}

// StringToBool turns a string into a bool
// The empty string is considered false
func StringToBool(s string) (bool, error) {
	if s == "" {
		return false, nil
	}

	b, err := strconv.ParseBool(s)
	if err != nil {
		return false, fmt.Errorf("failed to parse %s as bool", s)
	}

	return b, nil
}

// StringToInt interprets a string as an integer
func StringToInt(s string) (int, error) {
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("failed to parse %s as int", s)
	}
	return v, nil
}

// StringToUint8 interprets a strings an unsigned 8-bit integer
func StringToUint8(s string) (uint8, error) {
	v, err := strconv.ParseUint(s, 10, 8)
	if err != nil {
		return 0, fmt.Errorf("failed to parse %s as a uint8", s)
	}
	return uint8(v), nil
}

// StringToUint16 interprets a strings an unsigned 16-bit integer
func StringToUint16(s string) (uint16, error) {
	v, err := strconv.ParseUint(s, 10, 16)
	if err != nil {
		return 0, fmt.Errorf("failed to parse %s as a uint16", s)
	}
	return uint16(v), nil
}

// ToPhilipsBrightness converts between a Hemtjanst Brightness and a Philips Hue brightness
func ToPhilipsBrightness(i int) int {
	i = (i * 254) / 100
	if i == 0 {
		return 1
	}
	return i
}

// ToHemtjanstBrightness converts brightness to Hemtjanst brightness
func ToHemtjanstBrightness(i int) int {
	return (i * 100) / 254
}

// HemtjanstHStoCIExy takes a hue/saturation and turns it into CIE xy coordinates
func HemtjanstHStoCIExy(h, s int) []float64 {
	c := colorful.Hsv(float64(h), float64(s)/100, 1)
	x, y, _ := c.Clamped().Xyy()
	return []float64{x, y}
}

// CIExyToHemtjanstHS takes CIE xy and transforms them to HS
func CIExyToHemtjanstHS(x, y float64) (int, int) {
	c := colorful.Xyy(x, y, 1)
	h, s, _ := c.Clamped().Hsv()
	return int(h), int(s * 100)
}

// TopicToStrInt turns a topic name into a stringified integer. This should
// generate a stable identifier that can be used for the group and light keys.
//
// Though the iOS version of the Hue app has no issue with light and group
// keys not being a stringified integer, the Android version crashes.
func TopicToStrInt(topic string) string {
	sum := sha1.Sum([]byte(topic))
	return strconv.Itoa(
		int(binary.BigEndian.Uint32(append([]byte{sum[0] & 0x7F}, sum[1:4]...))),
	)
}

// BoolToStr takes a boolean and returns on/off
func BoolToStr(b bool) string {
	if b {
		return "on"
	}
	return "off"
}

// IntToStr stringifies an int
func IntToStr(i int) string {
	return strconv.Itoa(i)
}

// IntPtr returns a pointer to an int
func IntPtr(i int) *int {
	return &i
}

// StrPtr returns a pointer to a string
func StrPtr(s string) *string {
	return &s
}

// FloatPtr returns a pointer to a slice of float64s
func FloatPtr(f []float64) *[]float64 {
	return &f
}

// BoolPtr returns a pointer to a bool
func BoolPtr(b bool) *bool {
	return &b
}
