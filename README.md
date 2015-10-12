# Char
Char is planned to be a programming language similar to Go but with a syntax closer to Python.

Currently in a very early experimental state.

### Current Syntax*
##### *very subject to change

```
// An example Char file, this is a comment.
// Blocks are specified via indentation, and tabs are interpreted as four spaces.

use							// Similar to Go-style imports
	"fmt"
	"io"

main						// The name of the class
	StaticProp  int			// A public, static integer property of the main class,
							//   indicated by the leading capital letter.
	.memberProp string		// A string property of an instance of the main class,
							//   indicated by the leading dot.
	
	main()					// Function declaration, static due to no leading dot.
		var x				// All variables must be declared with var before use.
		var s = "hello"		// Variables can be initialized during declaration.
		x = 3				// Assignment
		x = x + 3			// Basic math
		var b = true		// Boolean
		b = b or false		// Keywords 'and' and 'or'
		
		go/										// Go blocks allow the embedding
		fmt.Println("Hello from Char!")			//   of Go code directly into Char
		/go										//   for the time being, and will
												//   hopefully be removed later.
```
