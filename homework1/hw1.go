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

type Mod func() (string, int)

func change(s string, n int) func() (string, int) {
    return func() (string, int) {
        return s, n
    }
}

func sandglass(params ...Mod) {
    settings := map[string]int{
        "size" : 15,
        "border_char" : '#',
        "border_color" : 3,
        "is_sand" : 1,      // 0 - песка нет, 1 (или любое другое число) - песок есть
        "sand_char" : '$',
        "sand_level" : 3,   // для sand_level (высота уровня песка) можно указать значения от 0 и выше, но при значениях
        "sand_color" : 4,   // больших, чем size - 2, песок будет просто занимать всё пространство внутри часов
    }
    for _, param := range params {
        s, n := param()
        settings[s] = n
    }
    fmt.Println(string(colors[settings["border_color"]]) + strings.Repeat(string(settings["border_char"]), settings["size"]))
    start := 1
    end := settings["size"] - 2
    for i := settings["size"] - 2; i >= 1; i-- {
        for j := 0; j < settings["size"]; j++ {
            switch {
            case j == start || j == end:
                fmt.Print(string(colors[settings["border_color"]]) + string(settings["border_char"]))
            case settings["is_sand"] != 0 && j > Min(start, end) && j < Max(start, end) && i <= settings["sand_level"]:
                fmt.Print(string(colors[settings["sand_color"]]) + string(settings["sand_char"]))
            default:
                fmt.Print(" ")
            }
        }
        fmt.Print("\n")
        start++
        end--
    }
    fmt.Println(string(colors[settings["border_color"]]) + strings.Repeat(string(settings["border_char"]), settings["size"]))
}

func main() {
    // теперь можно вызывать без параметров (с настройками по умолчанию)
    sandglass()
    // либо при вызове изменять некоторые настройки на выбор
    sandglass(change("size", 17), change("border_color", 7))
}
