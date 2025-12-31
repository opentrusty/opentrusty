package identity

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
)

// GenerateRandomAvatar returns a vibrant SVG based on a hash of the email using HSL color space.
func GenerateRandomAvatar(email string) string {
	hash := sha256.Sum256([]byte(email))

	// HSL (Hue, Saturation, Lightness)
	// Hue: 0-360, picked from hash
	hue := float64(int(hash[0])%360 + (int(hash[1]) << 8 % 360))
	if hue > 360 {
		hue = float64(int(hue) % 360)
	}

	// Use fixed saturation and lightness for "Boring Avatars" look (bright and harmonious)
	saturation := 0.70 // 70%
	lightness := 0.60  // 60%

	bgColor := hslToHex(hue, saturation, lightness)
	textColor := "#ffffff"

	// Initials from email prefix
	initial := "?"
	if len(email) > 0 {
		initial = string(email[0])
	}

	// Simple SVG with a circle and the initial
	svg := fmt.Sprintf(`<svg width="100" height="100" viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
  <rect width="100" height="100" fill="%s" />
  <text x="50" y="50" dy=".35em" fill="%s" font-family="sans-serif" font-size="50" text-anchor="middle" font-weight="bold">%s</text>
</svg>`, bgColor, textColor, strings.ToUpper(initial))

	return "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(svg))
}

// hslToHex converts HSL values to a hex color string
func hslToHex(h, s, l float64) string {
	r, g, b := hslToRgb(h/360, s, l)
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

func hslToRgb(h, s, l float64) (uint8, uint8, uint8) {
	var r, g, b float64
	if s == 0 {
		r, g, b = l, l, l
	} else {
		var q float64
		if l < 0.5 {
			q = l * (1 + s)
		} else {
			q = l + s - l*s
		}
		p := 2*l - q
		r = hueToRgb(p, q, h+1.0/3.0)
		g = hueToRgb(p, q, h)
		b = hueToRgb(p, q, h-1.0/3.0)
	}
	return uint8(r * 255), uint8(g * 255), uint8(b * 255)
}

func hueToRgb(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}
	if t < 1.0/6.0 {
		return p + (q-p)*6*t
	}
	if t < 1.0/2.0 {
		return q
	}
	if t < 2.0/3.0 {
		return p + (q-p)*(2.0/3.0-t)*6
	}
	return p
}
