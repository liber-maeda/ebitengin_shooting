package main

import (
	"image/png"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Game struct{}

var runes []rune
var mousebutton0 bool = false
var mousebutton1 bool = false
var posX int = 0
var posY int = 0
var posx float64 = 0.0
var posy float64 = 0.0
var shotFlag bool = false
var shotx float64 = 0.0
var shoty float64 = 0.0
var enemyposx float64 = 250.0
var enemyposy float64 = 320.0
var enemyDirection int = 0
var enemyshotFlag bool = false
var enemyshotx float64 = 0.0
var enemyshoty float64 = 0.0
var touchX int = 0
var touchY int = 0
var touchId int = 0
var touchIDs []ebiten.TouchID
var drawImage *ebiten.Image
var shotImage *ebiten.Image
var enemyImage *ebiten.Image
var enemyshotImage *ebiten.Image
var enemyDirectionTime int64
var isEnemyHit bool = false
var isMyHit bool = false
var myWidth float64
var myHeight float64
var shotWidth float64
var shotHeight float64
var enemyWidth float64
var enemyHeight float64
var enemyShotWidth float64
var enemyShotHeight float64
var score int = 0

func (g *Game) Update() error {
	// ゲームオーバーなので何もしないようにする
	if isMyHit {
		return nil
	}

	// 自機の移動
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		posy -= 0.5
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		posy += 0.5
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		posx += 0.5
	}
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		posx -= 0.5
	}

	// 自機の弾発射（一発のみ）
	if ebiten.IsKeyPressed(ebiten.KeySpace) && !shotFlag {
		shotFlag = true
		shotx = posx + 13.0
		shoty = posy + 25.0
	}
	if shotFlag {
		shoty += 1.0
	}
	// 弾が画面外になったら非表示にする
	if shotx > 640 || shoty > 480 {
		shotFlag = false
	}

	if !isEnemyHit {
		// 敵の移動（数秒同じ方向、左右ランダム）
		// あとで画面外に消えないようにする
		if enemyDirection == 1 {
			enemyposx -= 5
		} else if enemyDirection == 2 {
			enemyposx += 5
		}
		if enemyDirectionTime+3 < time.Now().Unix() {
			rand.Seed(time.Now().UnixNano())
			enemyDirection = rand.Intn(3)
		}

		// 敵の弾発射
		if !enemyshotFlag {
			enemyshotFlag = true
			enemyshotx = enemyposx + 60
			enemyshoty = enemyposy
		} else {
			enemyshoty -= 8.0
			// 範囲外に行ったら非表示
			if enemyshotx < 0 || enemyshoty < 0 {
				enemyshotFlag = false
			}
		}

		// 自機と敵　か　自機と敵の弾の当たり判定
		if isHit(posx, posy, myWidth, myHeight, enemyposx, enemyposy, enemyWidth, enemyHeight) ||
			isHit(posx, posy, myWidth, myHeight, enemyshotx, enemyshoty, enemyShotWidth, enemyShotHeight) {
			// 当たっているので自機消失（ゲームオーバー）
			isMyHit = true
		}
	}

	// 自機の弾と敵の当たり判定
	if isHit(shotx, shoty, shotWidth, shotHeight, enemyposx, enemyposy, enemyWidth, enemyHeight) {
		// 当たってるので敵消失
		isEnemyHit = true
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// 自機の弾
	if shotFlag {
		op2 := &ebiten.DrawImageOptions{}
		op2.GeoM.Scale(0.1, 0.1)
		op2.GeoM.Translate(shotx, shoty)
		screen.DrawImage(shotImage, op2)
		shotWidth = float64(shotImage.Bounds().Dx()) * 0.1
		shotHeight = float64(shotImage.Bounds().Dy()) * 0.1
	}

	if isEnemyHit {
		// 敵非表示
		ebitenutil.DebugPrint(screen, "enemyHit!!!")
		score += 100
		enemyWidth = 0
		enemyHeight = 0
		enemyShotWidth = 0
		enemyShotHeight = 0
	} else {
		//ebitenutil.DebugPrint(screen, fmt.Sprintf("%f:%f:%f:%f / %f:%f:%f:%f", shotx, shoty, shotWidth, shotHeight, enemyposx, enemyposy, enemyWidth, enemyHeight))
		// 敵の弾
		if enemyshotFlag {
			op4 := &ebiten.DrawImageOptions{}
			op4.GeoM.Scale(0.1, 0.1)
			op4.GeoM.Translate(enemyshotx, enemyshoty)
			//ebitenutil.DebugPrint(screen, fmt.Sprintf("%d:%d:%d:%d", enemyshotImage.Bounds().Dx(), enemyshotImage.Bounds().Dy(), enemyshotImage.Bounds().Size().X, enemyshotImage.Bounds().Size().Y))
			screen.DrawImage(enemyshotImage, op4)
			enemyShotWidth = float64(enemyshotImage.Bounds().Dx()) * 0.1
			enemyShotHeight = float64(enemyshotImage.Bounds().Dy()) * 0.1
		}
		// 敵
		op3 := &ebiten.DrawImageOptions{}
		op3.GeoM.Scale(0.3, 0.3)
		op3.GeoM.Translate(enemyposx, enemyposy)
		screen.DrawImage(enemyImage, op3)
		enemyWidth = float64(enemyImage.Bounds().Dx()) * 0.3
		enemyHeight = float64(enemyImage.Bounds().Dy()) * 0.3
	}
	// 自機
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(0.1, 0.1)
	op.GeoM.Translate(posx, posy)
	screen.DrawImage(drawImage, op)
	myWidth = float64(drawImage.Bounds().Dx()) * 0.1
	myHeight = float64(drawImage.Bounds().Dy()) * 0.1

	if isMyHit {
		ebitenutil.DebugPrint(screen, "GameOver!!!")
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}

func main() {
	/*
		// ファイル開く
		file, _ := os.Open("./quadcopter_drone.png")
		defer file.Close()
		// デコード
		img, err := png.Decode(file)
		if err != nil {
			panic(err)
		}
		drawImage = ebiten.NewImageFromImage(img)
	*/
	drawImage = imageOpen("./quadcopter_drone.png")

	/*
		// ファイルを開く
		file2, _ := os.Open("./ball11_gold.png")
		defer file2.Close()
		// デコード
		img2, err2 := png.Decode(file2)
		if err2 != nil {
			panic(err2)
		}
		shotImage = ebiten.NewImageFromImage(img2)
	*/
	shotImage = imageOpen("./ball11_gold.png")

	//
	enemyImage = imageOpen("./animal_hebi_cobra.png")

	//
	enemyshotImage = imageOpen("./ball01_red.png")

	enemyDirectionTime = time.Now().Unix()

	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Hello, World!")
	if err := ebiten.RunGame(&Game{}); err != nil {
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
