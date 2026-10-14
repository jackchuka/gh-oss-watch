# Regenerate with: vhs assets/demo.tape (from the repository root)
# Requires an authenticated gh CLI. Uses a throwaway HOME so your real watch list is untouched.

Output assets/demo.gif

Set Shell "bash"
Set FontSize 22
Set Width 1440
Set Height 760
Set Padding 24
Set Theme "Catppuccin Mocha"
Set TypingSpeed 45ms

Hide
Type `export GH_TOKEN="$(gh auth token)" D=/tmp/gh-oss-watch-demo && rm -rf "$D" && mkdir -p "$D" && go build -o "$D/gh-oss-watch" . && export HOME="$D" && cd "$D" && gh() { if [ "$1" = oss-watch ]; then shift; "$D/gh-oss-watch" "$@"; else command gh "$@"; fi; } && for r in mdschema gh-md goalias confluence-md; do gh oss-watch add "jackchuka/$r" stars issues pull_requests forks releases >/dev/null; done && export PS1='\[\e[1;35m\]$ \[\e[0m\]' && clear`
Enter
Wait@90s /^\$$/
Type "clear"
Enter
Sleep 1s
Show

Type "gh oss-watch add jackchuka/gh-oss-watch"
Sleep 400ms
Enter
Sleep 2s

Type "gh oss-watch list"
Sleep 400ms
Enter
Sleep 3s

Type "clear"
Enter
Type "gh oss-watch dashboard"
Sleep 400ms
Enter
Sleep 5s

Type "gh oss-watch releases"
Sleep 400ms
Enter
Sleep 5s
