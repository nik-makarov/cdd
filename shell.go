package main

import "strings"

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}

func shellInit(executable string) string {
	// The sentinel preserves trailing newlines in directory names, which shell
	// command substitution would otherwise strip. Paths never pass through eval.
	return `0go() {
    local _0go_arg _0go_dir _0go_from="$PWD"
    for _0go_arg in "$@"; do
        case "$_0go_arg" in
            --) break ;;
            -h|--help|--version|--init|--init=*)
                command ` + shellQuote(executable) + ` "$@"
                return "$?"
                ;;
        esac
    done
    _0go_dir="$(command ` + shellQuote(executable) + ` "$@" && printf '.')" || return "$?"
    _0go_dir="${_0go_dir%.}"
    [ -n "$_0go_dir" ] || return 0
    builtin cd -- "$_0go_dir" || return "$?"
    if [ "$_0go_from" != "$PWD" ]; then
        builtin printf '0go: %q -> %q\n' "$_0go_from" "$PWD" >&2
    fi
    return 0
}
`
}
