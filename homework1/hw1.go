package main

import (
    "fmt"
    "strings"
)

func Max(x, y int) int {
    if x < y {
        return y
    }
    return x
}

func Min(x, y int) int {
    if x > y {
        return y
    }
    return x
}

var colors = map[int]string {
    1 : "\033[0m", // reset color
    2 : "\033[31m", // red
    3 : "\033[32m", // green
    4 : "\033[33m", // yellow
    5 : "\033[34m", // blue
    6 : "\033[35m", // purple
    7 : "\033[36m", // cyan
    8 : "\033[37m", // white
}

func sandglass(params map[string]int) {
	fmt.Println(string(colors[params["border_color"]]) + strings.Repeat(string(params["border_char"]), params["size"]))
    start := 1
    end := params["size"] - 2
    for i := params["size"] - 2; i >= 1; i-- {
        for j := 0; j < params["size"]; j++ {
            switch {
            case j == start || j == end:
                fmt.Print(string(colors[params["border_color"]]) + string(params["border_char"]))
            case params["is_sand"] != 0 && j > Min(start, end) && j < Max(start, end) && i <= params["sand_level"]:
                fmt.Print(string(colors[params["sand_color"]]) + string(params["sand_char"]))
            default:
                fmt.Print(" ")
            }
        }
        fmt.Print("\n")
        start++
        end--
    }
    fmt.Println(string(colors[params["border_color"]]) + strings.Repeat(string(params["border_char"]), params["size"]))
}

func main() {
    // первые 4 параметра обязательные; остальные используются, если в "is_sand" указать 1,
    // иначе их можно не задавать
    settings := map[string]int{
        "size" : 15,
        "border_char" : '#',
        "border_color" : 3,
        "is_sand" : 1,      // 0 - песка нет, 1 (или любое другое число) - песок есть
        "sand_char" : '$',
        "sand_level" : 3,   // для sand_level (высота уровня песка) можно указать значения от 0 и выше, но при значениях
        "sand_color" : 4,   // больших, чем size - 2, песок будет просто занимать всё пространство внутри часов
    }
	sandglass(settings)
}