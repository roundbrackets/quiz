/* Check if a string has been rotated. Think rot13. */

package main
import "fmt"

func main() {

    comp_str := "abcdefgh"
    strings := []string{ "cdefghab", "abcdefgh", "abcdefg", "abcdexfgh","habcdefgh" }

    fmt.Println("hello")

    for _, str := range strings {
        if true == isRotated(comp_str, str) {
            fmt.Println("yes!")
        } else{
            fmt.Println("false!")
        }
    }
}

func isRotated (haystack string, needle string) (bool) {
    if haystack == needle {
        return false
    }

    for i, _ := range haystack {
        head := haystack[0:i]
        tail := haystack[i:len(haystack)]

        fmt.Printf("%s %s %s %s == %s\n", i, haystack, tail, head, needle)

        if tail + head == needle {

            fmt.Printf("TRUE %s %s %s %s == %s\n", i, haystack, tail, head, needle)

            return true
        }
    }

    return false

}
