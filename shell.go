package main

import "strings"

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}

func shellSetupMessage(styles viewStyles) string {
	heading := "cdd: shell integration is not loaded."
	commands := `echo 'eval "$(/opt/homebrew/bin/cdd --init zsh)"' >> ~/.zshrc
source ~/.zshrc`
	if styles.failure != "" {
		heading = styles.failure + heading + "\x1b[0m"
	}
	if styles.accent != "" {
		commands = styles.accent + commands + "\x1b[0m"
	}
	return heading + "\nFor zsh on Apple Silicon macOS:\n\n" + commands + `

The configuration line enables cdd in future zsh sessions; source activates it
in the current terminal. Run the echo command once, then close and reopen your terminal.

On Intel macOS or Linux, replace /opt/homebrew with the output of
brew --prefix. For bash, use --init bash in ~/.bashrc.
`
}

func shellInit(executable string) string {
	// The sentinel preserves trailing newlines in directory names, which shell
	// command substitution would otherwise strip. Paths never pass through eval.
	// Only selected paths begin with /; relay informational output without cd.
	return `cdd() {
    local _cdd_dir _cdd_from="$PWD"
    _cdd_dir="$(CDD_SHELL_INTEGRATION=1 command ` + shellQuote(executable) + ` "$@" && printf '.')" || return "$?"
    _cdd_dir="${_cdd_dir%.}"
    [ -n "$_cdd_dir" ] || return 0
    case "$_cdd_dir" in
        /*) ;;
        *) builtin printf '%s' "$_cdd_dir"; return "$?" ;;
    esac
    builtin cd -- "$_cdd_dir" || return "$?"
    if [ "$_cdd_from" != "$PWD" ]; then
        builtin printf 'cdd: %q -> %q\n' "$_cdd_from" "$PWD" >&2
    fi
    return 0
}
`
}
