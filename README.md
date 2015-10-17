# Char
Char is planned to be a programming language similar to Go but with a syntax closer to Python.

Currently in a very early experimental state.

### Current Syntax*
###### *very subject to change

```
// An example Char file, this is a comment.
// Blocks are specified via indentation, and tabs are interpreted as four spaces.

use							// Similar to Go-style imports
	"fmt"
	"io"

main						// The name of the class
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

// A public class due to the first letter being uppercase
MyClass
	const									// Class constants
		greeting = "Hello from Char!"		// String constant
		First = iota						// Like Go, iota starts at 0 per const
		Second								//   block and constants without an
		Third								//   assignment use the prior one.
	
	privateProp int				// A private property since first letter is lowercase.
	PublicProp  string			// A public property since first letter is uppercase.
	
	print()							// A static function, referenced by MyClass.print()
		var s = MyClass.greeting	// Getting a constant. Only done here because we
		go/							//   have to use a Go block right now to print.
		fmt.Println(s)
		/go
	
	.Add(v1, v2 int) int				// Starting dot indicates a method. Parameters and
		privateProp += v1 * v2			//   return types specified the same as Go.
		return this.privateProp			//   Can use an implicit or explicit 'this'
```
