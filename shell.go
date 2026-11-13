package main

import "strings"

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}

func shellInit(executable string) string {
	// The sentinel preserves trailing newlines in directory names, which shell
	// command substitution would otherwise strip. Paths never pass through eval.
	return `cdd() {
    local _cdd_arg _cdd_dir _cdd_from="$PWD"
    for _cdd_arg in "$@"; do
        case "$_cdd_arg" in
            --) break ;;
            -h|--help|--version|--init|--init=*)
                command ` + shellQuote(executable) + ` "$@"
                return "$?"
                ;;
        esac
    done
    _cdd_dir="$(command ` + shellQuote(executable) + ` "$@" && printf '.')" || return "$?"
    _cdd_dir="${_cdd_dir%.}"
    [ -n "$_cdd_dir" ] || return 0
    builtin cd -- "$_cdd_dir" || return "$?"
    if [ "$_cdd_from" != "$PWD" ]; then
        builtin printf 'cdd: %q -> %q\n' "$_cdd_from" "$PWD" >&2
    fi
    return 0
}
`
}
