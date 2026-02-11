# demo.tape - Full tgdash demo
# Covers: exec mode, dashboard view, navigation, scrolling,
# list view, error filtering

Output site/static/demo/demo.gif

Set Shell bash
Set FontSize 28
Set Width 1600
Set Height 800
Set Padding 30
Set Theme "Catppuccin Mocha"
Set TypingSpeed 50ms

# Build tgdash and set clean prompt
Hide
Type `export PS1="$ "`
Enter
Type "go build -o /tmp/tgdash ."
Enter
Sleep 3s
Type "clear"
Enter
Sleep 500ms
Show

# Launch with exec mode
Type "/tmp/tgdash -- cat internal/integration/testdata/terragrunt_new_format.txt"
Sleep 1s
Enter

# Watch the dashboard render and units progress
Sleep 5s

# Navigate down through units
Down
Sleep 500ms
Down
Sleep 500ms
Down
Sleep 500ms

# Navigate back up
Up
Sleep 500ms
Up
Sleep 500ms

# Scroll the output viewport with j/k
Type "j"
Sleep 500ms
Type "j"
Sleep 500ms
Type "k"
Sleep 500ms

# Switch to list view
Type "	"
Sleep 2s

# Switch back to dashboard
Type "	"
Sleep 2s

# Jump to bottom with G
Type "G"
Sleep 1s

# Jump to top with gg
Type "gg"
Sleep 1s

# Toggle error filter
Type "e"
Sleep 2s

# Toggle back
Type "e"
Sleep 2s

# Exit
Type "q"
