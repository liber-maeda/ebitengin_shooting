package main

import (
	"fmt"
	"image/png"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const ENEMY_MAX = 3
const SHOT_MAX = 3

var runes []rune
var mousebutton0 bool = false
var mousebutton1 bool = false

// var posX int = 0
// var posY int = 0
// var posx float64 = 0.0
// var posy float64 = 0.0
// var shotFlag bool = false
// var shotx float64 = 0.0
// var shoty float64 = 0.0
// var enemyposx float64 = 250.0
// var enemyposy float64 = 320.0
var enemyDirection int = 0

// var enemyshotFlag bool = false
// var enemyshotx float64 = 0.0
// var enemyshoty float64 = 0.0
var touchX int = 0
var touchY int = 0
var touchId int = 0
var touchIDs []ebiten.TouchID

// var drawImage *ebiten.Image
// var shotImage *ebiten.Image
// var enemyImage *ebiten.Image
// var enemyshotImage *ebiten.Image
var enemyDirectionTime int64

// var isEnemyHit bool = false
// var isMyHit bool = false
// var myWidth float64
// var myHeight float64
// var shotWidth float64
// var shotHeight float64
// var enemyWidth float64
// var enemyHeight float64
// var enemyShotWidth float64
// var enemyShotHeight float64

type baseData struct {
	posx        float64
	posy        float64
	isHit       bool
	drawImage   *ebiten.Image
	drawWidth   float64
	drawHeight  float64
	shot        [SHOT_MAX]shotData
	revivalTime time.Time
}

type shotData struct {
	shotx      float64
	shoty      float64
	shotFlag   bool
	drawImage  *ebiten.Image
	drawWidth  float64
	drawHeight float64
}

type Game struct {
	my    baseData
	enemy [ENEMY_MAX]baseData
	score int
}

func NewGame() *Game {
	// Create the game
	game := &Game{}
	game.Init()
	return game
}

func (g *Game) Init() {
	// 初期処理

	g.my.posx = 0.0
	g.my.posy = 0.0
	g.my.isHit = false
	// 画像ロード
	g.my.drawImage = imageOpen("./quadcopter_drone.png")
	for i := 0; i < SHOT_MAX; i++ {
		g.my.shot[i].drawImage = imageOpen("./ball11_gold.png")
		g.my.shot[i].shotFlag = false
	}

	for i := 0; i < ENEMY_MAX; i++ {
		g.enemy[i].posx = 150.0 + float64(100.0*i)
		g.enemy[i].posy = 320.0
		g.enemy[i].isHit = false
		// 画像ロード
		g.enemy[i].drawImage = imageOpen("./animal_hebi_cobra.png")
		for j := 0; j < SHOT_MAX; j++ {
			g.enemy[i].shot[j].drawImage = imageOpen("./ball01_red.png")
			g.enemy[i].shot[j].shotFlag = false
		}
	}

	g.score = 0
}

func (g *Game) Update() error {
	// ゲームオーバーなので何もしないようにする
	if g.my.isHit {
		return nil
	}

	// 自機の移動
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		if g.my.posy >= 0.0 {
			g.my.posy -= 0.5
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		if g.my.posy <= 480.0-g.my.drawHeight {
			g.my.posy += 0.5
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		if g.my.posx >= 0.0 {
			g.my.posx -= 0.5
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		if g.my.posx <= 640.0-g.my.drawWidth {
			g.my.posx += 0.5
		}
	}

	// 自機の弾発射
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		//if ebiten.IsKeyPressed(ebiten.KeySpace) {
		for i := 0; i < SHOT_MAX; i++ {
			if !g.my.shot[i].shotFlag {
				g.my.shot[i].shotFlag = true
				g.my.shot[i].shotx = g.my.posx + 13.0
				g.my.shot[i].shoty = g.my.posy + 25.0
				break
			}
		}
	}
	for i := 0; i < SHOT_MAX; i++ {
		// 弾の自動移動
		if g.my.shot[i].shotFlag {
			g.my.shot[i].shoty += 1.0
		}
		// 弾が画面外になったら非表示にする
		if g.my.shot[i].shotx > 640 || g.my.shot[i].shoty > 480 {
			g.my.shot[i].shotFlag = false
		}
	}

	for i := 0; i < ENEMY_MAX; i++ {
		if !g.enemy[i].isHit {
			// 敵の移動（数秒同じ方向、左右ランダム）
			if enemyDirection == 1 {
				if g.enemy[i].posx > 0.0 {
					g.enemy[i].posx -= 5
				}
			} else if enemyDirection == 2 {
				if g.enemy[i].posx < 640.0-g.enemy[i].drawWidth {
					g.enemy[i].posx += 5
				}
			}
			if enemyDirectionTime+3 < time.Now().Unix() {
				rand.Seed(time.Now().UnixNano())
				enemyDirection = rand.Intn(3)
			}

			// 敵の弾発射
			for j := 0; j < SHOT_MAX; j++ {
				if !g.enemy[i].shot[j].shotFlag {
					g.enemy[i].shot[j].shotFlag = true
					g.enemy[i].shot[j].shotx = g.enemy[i].posx + 60
					g.enemy[i].shot[j].shoty = g.enemy[i].posy
				} else {
					g.enemy[i].shot[j].shoty -= 8.0
					// 範囲外に行ったら非表示
					if g.enemy[i].shot[j].shotx < 0 || g.enemy[i].shot[j].shoty < 0 {
						g.enemy[i].shot[j].shotFlag = false
					}
				}

				// 自機と敵　か　自機と敵の弾の当たり判定
				if isHit(g.my.posx, g.my.posy, g.my.drawHeight, g.my.drawHeight, g.enemy[i].posx, g.enemy[i].posy, g.enemy[i].drawWidth, g.enemy[i].drawHeight) ||
					isHit(g.my.posx, g.my.posy, g.my.drawWidth, g.my.drawHeight, g.enemy[i].shot[j].shotx, g.enemy[i].shot[j].shoty, g.enemy[i].shot[j].drawWidth, g.enemy[i].shot[j].drawHeight) {
					// 当たっているので自機消失（ゲームオーバー）
					g.my.isHit = true
				}

				// 自機の弾と敵の当たり判定（敵と自機の弾の数が一緒なので同じループで実施）
				if isHit(g.my.shot[j].shotx, g.my.shot[j].shoty, g.my.shot[j].drawWidth, g.my.shot[j].drawHeight, g.enemy[i].posx, g.enemy[i].posy, g.enemy[i].drawWidth, g.enemy[i].drawHeight) {
					// 当たってるので敵消失
					g.enemy[i].isHit = true
					g.score += 100
					g.my.shot[j].shotFlag = false
					//g.my.shot[j].shotx = g.my.posx
					//g.my.shot[j].shoty = g.my.posy
					t := time.Now()
					g.enemy[i].revivalTime = t.Add(time.Second + 10)
				}
			}
		} else {
			//時間経ってたら復活
			if g.enemy[i].revivalTime.Before(time.Now()) {
				g.enemy[i].isHit = false
			}
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// 自機の弾
	for i := 0; i < SHOT_MAX; i++ {
		if g.my.shot[i].shotFlag {
			op2 := &ebiten.DrawImageOptions{}
			op2.GeoM.Scale(0.1, 0.1)
			op2.GeoM.Translate(g.my.shot[i].shotx, g.my.shot[i].shoty)
			screen.DrawImage(g.my.shot[i].drawImage, op2)
			g.my.shot[i].drawWidth = float64(g.my.shot[i].drawImage.Bounds().Dx()) * 0.1
			g.my.shot[i].drawHeight = float64(g.my.shot[i].drawImage.Bounds().Dy()) * 0.1
		}
	}

	for i := 0; i < ENEMY_MAX; i++ {
		if g.enemy[i].isHit {
			// 敵非表示
			//ebitenutil.DebugPrint(screen, "enemyHit!!!")
			g.enemy[i].drawWidth = 0
			g.enemy[i].drawHeight = 0
			for j := 0; j < SHOT_MAX; j++ {
				g.enemy[i].shot[j].drawWidth = 0
				g.enemy[i].shot[j].drawHeight = 0
			}
		} else {
			//ebitenutil.DebugPrint(screen, fmt.Sprintf("%f:%f:%f:%f / %f:%f:%f:%f", shotx, shoty, shotWidth, shotHeight, enemyposx, enemyposy, enemyWidth, enemyHeight))
			for j := 0; j < SHOT_MAX; j++ {
				// 敵の弾
				if g.enemy[i].shot[j].shotFlag {
					op4 := &ebiten.DrawImageOptions{}
					op4.GeoM.Scale(0.1, 0.1)
					op4.GeoM.Translate(g.enemy[i].shot[j].shotx, g.enemy[i].shot[j].shoty)
					//ebitenutil.DebugPrint(screen, fmt.Sprintf("%d:%d:%d:%d", enemyshotImage.Bounds().Dx(), enemyshotImage.Bounds().Dy(), enemyshotImage.Bounds().Size().X, enemyshotImage.Bounds().Size().Y))
					screen.DrawImage(g.enemy[i].shot[j].drawImage, op4)
					g.enemy[i].shot[j].drawWidth = float64(g.enemy[i].shot[j].drawImage.Bounds().Dx()) * 0.1
					g.enemy[i].shot[j].drawHeight = float64(g.enemy[i].shot[j].drawImage.Bounds().Dy()) * 0.1
				}
			}
			// 敵
			op3 := &ebiten.DrawImageOptions{}
			op3.GeoM.Scale(0.3, 0.3)
			op3.GeoM.Translate(g.enemy[i].posx, g.enemy[i].posy)
			screen.DrawImage(g.enemy[i].drawImage, op3)
			g.enemy[i].drawWidth = float64(g.enemy[i].drawImage.Bounds().Dx()) * 0.3
			g.enemy[i].drawHeight = float64(g.enemy[i].drawImage.Bounds().Dy()) * 0.3
		}
	}
	// 自機
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(0.1, 0.1)
	op.GeoM.Translate(g.my.posx, g.my.posy)
	screen.DrawImage(g.my.drawImage, op)
	g.my.drawWidth = float64(g.my.drawImage.Bounds().Dx()) * 0.1
	g.my.drawHeight = float64(g.my.drawImage.Bounds().Dy()) * 0.1

	ebitenutil.DebugPrint(screen, fmt.Sprintf("score:%d", g.score))
	if g.my.isHit {
		ebitenutil.DebugPrint(screen, fmt.Sprintf("GameOver!!!:%d", g.score))
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}

func main() {
	enemyDirectionTime = time.Now().Unix()

	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Hello, World!")

	game := NewGame()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

func imageOpen(name string) *ebiten.Image {
	// ファイルオープン
	file, _ := os.Open(name)
	defer file.Close()
	// デコード
	img, err := png.Decode(file)
	if err != nil {
		panic(err)
	}
	return ebiten.NewImageFromImage(img)
}

func isHit(posx1 float64, posy1 float64, width1 float64, height1 float64, posx2 float64, posy2 float64, width2 float64, height2 float64) bool {
	// x座標に関する判定
	if posx2+width2 < posx1 {
		return false
	}
	if posx1+width1 < posx2 {
		return false
	}

	// y座標に関する判定
	if posy2+height2 < posy1 {
		return false
	}
	if posy1+height1 < posy2 {
		return false
	}

	return true
}
