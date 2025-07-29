# TODO

Here’s a list of **practical refinements** to make your REPL more usable and helpful for real users of your DSL.

---

## ✅ REPL Refinements: Immediate Usability

### 1. **Line Continuation Prompt**

Switch prompt based on context:

```go
if len(lines) > 0 {
	fmt.Print(". ") // continuation line
} else {
	fmt.Print("> ")
}
```

---

### 2. **Better Block Detection**

Improve `blockComplete()` to:

* Track actual `if`/`for`/`do`/`then`/`end` balance.
* Avoid counting `if` in `"string literals"`.

Eventually, parse a partial AST and catch a recoverable `EOF` error.

---

### 3. **Result Inspection**

Add built-ins or DSL keywords for inspecting state:

```pascal
:vars       // shows declared variables
:hexes      // prints map.hexes
:dump       // show all map state
```

Then in your REPL driver:

```go
switch strings.TrimSpace(line) {
case ":vars":
	dumpVars(vm)
	continue
case ":hexes":
	dumpHexes(vm.root.Hexes)
	continue
}
```

---

### 4. **History & Editing**

Integrate [`github.com/chzyer/readline`](https://github.com/chzyer/readline):

* Input history with ↑↓
* Ctrl+R reverse search
* Multi-line input support
* Emacs/vi keybindings

Very little setup is required and makes a huge difference.

---

### 5. **Command Preprocessing**

Prefix REPL commands with `:` to avoid ambiguity:

| Input           | Action                      |
| --------------- | --------------------------- |
| `:exit`         | quit                        |
| `:help`         | print commands              |
| `:save map.wxx` | save to disk (mock or real) |

Use `strings.HasPrefix(line, ":")` to route these.

---

### 6. **Graceful Error Reporting**

* Wrap all execution in `defer recover()` to catch panics and print a helpful message.
* Print stack trace (optional) in dev mode.

Example:

```go
defer func() {
	if r := recover(); r != nil {
		fmt.Println("Parser panic:", r)
	}
}()
```

---

## 🔄 Stateful Features

### 7. **Define Global Variables**

Allow things like:

```pascal
x := 5;
```

Store `x` in `vm.vars`, retrievable later.

---

### 8. **Persist/Load WXX**

Once real XML support is added:

```pascal
load("foo.wxx");
save("bar.wxx");
```

Built-ins would map to:

```go
"load": func(args []Value) error {
	return vm.LoadFromFile(args[0].(string))
}
```

---

### 9. **Pretty-Print Map**

Add something like:

```pascal
print map.hexes;
```

In Go, make a helper:

```go
func dumpHexes(hexes []Hex) {
	for i, h := range hexes {
		fmt.Printf("[%d]: %s\n", i, h.Terrain)
	}
}
```

---

### 10. **Debug Flag**

Support `--debug` or `--trace` flags:

* Echo tokens
* Print AST
* Trace VM execution (optional)

---

## 🧰 Optional Extras for Later

| Feature                    | Benefit                            |
| -------------------------- | ---------------------------------- |
| AST visualization          | Great for debugging the DSL        |
| Auto-completion (readline) | Helps users explore functions/vars |
| Command aliasing           | E.g. `:q → :exit`                  |
| DSL documentation command  | `:doc print`                       |

---
