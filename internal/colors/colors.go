package colors

// Ansi colors
const (
	red    = "\033[31m"
	yellow = "\033[93m"
	pink   = "\033[95m"
	cyan   = "\033[96m"
	green  = "\033[92m"
	blue   = "\033[96m"
	reset  = "\033[0m"
)

func Red(text string) string {
	return red + text + reset
}

func Yellow(text string) string {
	return yellow + text + reset
}

func Green(text string) string {
	return green + text + reset
}

func Blue(text string) string {
	return blue + text + reset
}

func Pink(text string) string {
	return pink + text + reset
}

func Cyan(text string) string {
	return cyan + text + reset
}
