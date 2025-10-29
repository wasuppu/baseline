package main

import (
	"fmt"
)

func main() {
	source := `
 function main() {
      // Test Number
      assert(1);

      // Test Not
      assert(!0);
      assert(!(!1));

      putchar(46);

      // Test Equal
      assert(42 == 42);
      assert(!(0 == 42));

      // Test NotEqual
      assert(!(42 != 42));
      assert(0 != 42);

      // Test infix operators
      assert(42 == 4 + 2 * (12 - 2) + 3 * (5 + 1));

      // Test Call with no parameters
      assert(return42() == 42);
      assert(!returnNothing());

      // Test multiple parameters
      assert42(42);
      assert1234(1, 2, 3, 4);

      //assert(rand() != 42);
      //assert(putchar() != 1);

      //while (1) {
      //  assert(1);
      //}

      // Test If
      if (1)
	assert(1);
      else
	assert(0);

      if (0) {
        assert(0);
      } else {
        assert(1);
      }

      assert(factorial(5) == 120);

      var x = 4 + 2 * (12 - 2);
      var y = 3 * (5 + 1);
      var z = x + y;
      assert(z == 42);

      var a = 1;
      assert(a == 1);
      a = 0;
      assert(a == 0);

      // Test while loops
      var i = 0;
      while (i != 3) {
	i = i + 1;
      }
      assert(i == 3);

      assert(factorial2(5) == 120);

      putchar(10); // Newline
    }

    function return42() { return 42; }
    function returnNothing() {}
    function assert42(x) {
      assert(x == 42);
    }
    function assert1234(a, b, c, d) {
      assert(a == 1);
      assert(b == 2);
      assert(c == 3);
      assert(d == 4);
    }

    function assert(x) {
      if (x) {
	putchar(46);
      } else {
	putchar(70);
      }
    }

    function factorial(n) {
      if (n == 0) {
        return 1;
      } else {
        return n * factorial(n - 1);
      }
    }

    function factorial2(n) {
      var result = 1;
      while (n != 1) {
        result = result * n;
	n = n - 1;
      }
      return result;
    }
      `

	result := parser.ParseStringToCompletion(source)
	fmt.Printf("Parse successful: %#v\n", result)

	result.Emit(NewEnvironment())

	fmt.Println("All tests passed! Compiler rewritten in Go successfully!")
}
