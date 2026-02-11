# navigation.tape - Show keybindings and view switching
# Demonstrates: arrow keys for unit navigation, j/k for
# viewport scrolling, Tab to switch views

Output site/static/demo/navigation.gif

Set Shell bash
Set FontSize 14
Set Width 1200
Set Height 600
Set Theme "Catppuccin Mocha"
Set TypingSpeed 50ms

# Build and set clean prompt
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

Type "cat internal/integration/testdata/terragrunt_new_format.txt | /tmp/tgdash"
Sleep 1s
Enter

# Let it render
Sleep 3s

# Navigate down between units with arrow keys
Down
Sleep 500ms
Down
Sleep 500ms
Down
Sleep 500ms

# Navigate back up with arrow keys
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

# Switch to list view with Tab
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
Sleep 2s

# Exit
Type "q"
