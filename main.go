package main

import (
	"bufio"
	"fmt"
	"github.com/inancgumus/screen"
	"github.com/mattn/go-tty"
	"log"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"
)

type positionableI interface {
	getPosition() vec2i
}

type vec2i struct {
	x int
	y int
}

type snakeBody struct {
	pos   vec2i
	speed vec2i
	next  *snakeBody
}

type snake struct {
	pos   vec2i
	speed vec2i
	next  *snakeBody
	size  int
}

type food struct {
	pos vec2i
}

func (s *snake) getPosition() vec2i {
	return s.pos
}

func (f *food) getPosition() vec2i {
	return f.pos
}

/**
Рисует еду
*/
func (f *food) Draw(buffer *[]byte, screenSize vec2i) {
	foodPos := getLinearPos(f.pos, screenSize)
	(*buffer)[foodPos] = '8'
}

/**
Инициализирует змейку
*/
func (s *snake) init(pos vec2i, next []*snakeBody) {
	s.pos = pos
	if len(next) > 0 {
		s.next = next[0]
		next = next[1:]
		current := s.next
		for _, val := range next {
			current.next = val
			current = current.next
		}
	}
}

/**
Генерирует новую еду
*/
func generateRandomFood(snake *snake, screenSize vec2i) food {
	snakeMap := make(map[vec2i]bool)
	snakeMap[snake.pos] = true

	bodyPeace := snake.next
	for bodyPeace != nil {
		snakeMap[bodyPeace.pos] = true
		bodyPeace = bodyPeace.next
	}

	newPos := vec2i{rand.Intn(screenSize.x), rand.Intn(screenSize.y)}
	_, ok := snakeMap[newPos]
	for ok == true {
		newPos = vec2i{rand.Intn(screenSize.x), rand.Intn(screenSize.y)}
		_, ok = snakeMap[newPos]
	}

	return food{pos: newPos}
}

/**
Рисует кадр
*/
func renderFrame(buffer []byte, screenWriter *bufio.Writer) {
	screen.Clear()
	screen.MoveTopLeft()
	screenWriter.Write(buffer[:(len(buffer) - 1)])
	screenWriter.Flush()
}

func newEmptyBuffer(size int) []byte {
	return make([]byte, size, size)
}

/**
Рисует змейку
*/
func (s *snake) Draw(buffer *[]byte, screenSize vec2i) {

	snakeHeadPos := getLinearPos(s.pos, screenSize) // Позиция головы змеи в линейном массиве
	(*buffer)[snakeHeadPos] = 'Q'

	bodyPeace := s.next
	for bodyPeace != nil {
		bodyPeacePos := getLinearPos(bodyPeace.pos, screenSize) // Позиция части тела
		(*buffer)[bodyPeacePos] = '@'
		bodyPeace = bodyPeace.next
	}
}

/**
Двигает змейку
*/
func (s *snake) Move(speed vec2i, screenSize vec2i, appendNewPeace bool) bool {
	oldPos := s.pos
	s.speed = speed

	s.pos.x = (s.pos.x + s.speed.x + screenSize.x) % screenSize.x
	s.pos.y = (s.pos.y + s.speed.y + screenSize.y) % screenSize.y

	var prevPeace *snakeBody = nil
	var lastPos vec2i
	var oldSpeed vec2i
	bodyPeace := s.next
	for bodyPeace != nil {
		prevPeace = bodyPeace
		lastPos = bodyPeace.pos
		oldSpeed = bodyPeace.speed

		bodyPeace.pos = oldPos
		oldPos = lastPos

		if bodyPeace.pos.x == s.pos.x && bodyPeace.pos.y == s.pos.y {
			return true
		}

		bodyPeace = bodyPeace.next
	}
	if appendNewPeace {
		s.size += 1
		prevPeace.next = &snakeBody{
			pos:   lastPos,
			speed: oldSpeed,
		}
	}
	return false
}

func (s *snake) Eat(food *food) {
	// Poka nichego ne pridumal
}

/**
Есть ли коллизия
*/
func getCollision(object1 positionableI, object2 positionableI) bool {
	pos1 := object1.getPosition()
	pos2 := object2.getPosition()
	if pos1.x == pos2.x && pos1.y == pos2.y {
		return true
	}
	return false
}

/**
Функция возвращает позицию элемента двумерного массива в одномерном массиве
*/
func getLinearPos(pos vec2i, size vec2i) int {
	return pos.y*size.x + pos.x
}

/**
Функция которая возвращает код нажатой на данной момент кнопки
*/
func keyboardPolling(key *rune, mutex *sync.Mutex) {
	tty, err := tty.Open()
	if err != nil {
		log.Fatal(err)
	}
	var r rune
	var err1 error
	for {
		r, err1 = tty.ReadRune()
		if err1 != nil {
			log.Fatal(err1)
		}
		*key = r
	}
}

/**
Получает направление движения в зависимости от нажатой кнопки
*/
func getMovementVector(keyPressed rune) vec2i {
	MoveLeft := vec2i{-1, 0} // 97  <
	MoveRight := vec2i{1, 0} // 100 >
	MoveUp := vec2i{0, -1}   // 119 ^
	MoveDown := vec2i{0, 1}  // 115 v
	if keyPressed == 97 {
		return MoveLeft
	}
	if keyPressed == 100 {
		return MoveRight
	}
	if keyPressed == 119 {
		return MoveUp
	}
	if keyPressed == 115 {
		return MoveDown
	}
	return MoveUp
}

func drawHighScore(highScore int, buffer *[]byte, screenSize vec2i) {
	var center vec2i = vec2i{screenSize.x, screenSize.y}
	str := "You score: " + strconv.Itoa(highScore)
	center.x -= len(str) / 2
	for i, letter := range str {
		if letter < 128 {
			(*buffer)[getLinearPos(vec2i{center.x + i, center.y}, screenSize)] = byte(letter)
		}
	}
}

func main() {
	var keyPressed rune
	var keyMutex sync.Mutex
	rand.Seed(time.Now().UnixNano())

	var size vec2i
	size.x, size.y = screen.Size()
	center := vec2i{x: size.x / 2, y: size.y / 2}

	screenBuffer := bufio.NewWriter(os.Stdout)

	snake := snake{
		pos: center,
		next: &snakeBody{
			pos: vec2i{
				x: center.x - 1,
				y: center.y,
			},
		},
		size: 2,
	}

	food := generateRandomFood(&snake, size)

	var lastMovementKey rune = 97
	var keyCode rune

	// Считывание ввода с клавиатуры
	go keyboardPolling(&keyPressed, &keyMutex)

	var appendNewBodyPeace bool

	var highScore int = 0
	// Game loop
	for {
		size.x, size.y = screen.Size()
		bufferSize := size.x * size.y
		appendNewBodyPeace = false
		frameBuffer := newEmptyBuffer(bufferSize)
		// Определяем какая кнопка была нажата последней
		keyCode = keyPressed
		if keyCode == 97 || keyCode == 100 || keyCode == 119 || keyCode == 115 {
			lastMovementKey = keyCode
		}

		if getCollision(&snake, &food) {
			snake.Eat(&food)
			food = generateRandomFood(&snake, size)
			appendNewBodyPeace = true
		}
		food.Draw(&frameBuffer, size)

		// Игра закончилась
		if snake.Move(getMovementVector(lastMovementKey), size, appendNewBodyPeace) {
			highScore = snake.size
			break
		}

		snake.Draw(&frameBuffer, size)
		renderFrame(frameBuffer, screenBuffer)
		time.Sleep(250 * time.Millisecond)
	}

	screen.Clear()
	screen.MoveTopLeft()
	fmt.Println("You score: ", highScore)
	fmt.Scanln()
}
