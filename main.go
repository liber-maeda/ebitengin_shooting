package main

import (
	"image"
	"image/png"
	"io"
	"log"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	//"github.com/hajimehoshi/ebiten/v2/text/v2"
)

const (
	ENEMY_MAX           = 3
	SHOT_MAX            = 3
	EFFECT_FRAMEOX      = 4
	EFFECT_FRAMEOY      = 4
	EFFECT_FRAME_WIDTH  = 340
	EFFECT_FRAME_HEIGHT = 340
	EFFECT_FRAME_COUNT  = 9
	SCORE_OX            = 10
	SCORE_OY            = 0
	SCORE_WIDTH         = 32
	SCORE_HEIGHT        = 64
)

//var runes []rune
//var mousebutton0 bool = false
//var mousebutton1 bool = false

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
	effectImage *ebiten.Image
	effectCount int
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
	my               baseData
	enemy            [ENEMY_MAX]baseData
	score            int
	audioContext     *audio.Context
	shotSEBytes      []byte
	explosionSEBytes []byte
	scoreImage       *ebiten.Image
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
	g.my.effectImage = imageOpen("./Effect.png")
	g.my.effectCount = 0

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
		g.enemy[i].effectImage = imageOpen("./Effect.png")
		g.enemy[i].effectCount = 0
	}

	g.score = 0

	// 音設定
	g.audioContext = audio.NewContext(44100)
	g.shotSEBytes, _ = openSE("./shot.mp3")
	g.explosionSEBytes, _ = openSE("./explosion.mp3")

	// 得点用の画像ロード
	g.scoreImage = imageOpen("./score.png")
}

func (g *Game) Update() error {
	// ゲームオーバーなので何もしないようにする
	if g.my.isHit {
		// エフェクトのカウント
		g.my.effectCount++
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
				// 発射音
				sePlayer := g.audioContext.NewPlayerFromBytes(g.shotSEBytes)
				sePlayer.Play()
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
					g.enemy[i].posx -= 1
				}
			} else if enemyDirection == 2 {
				if g.enemy[i].posx < 640.0-g.enemy[i].drawWidth {
					g.enemy[i].posx += 1
				}
			}
			if enemyDirectionTime+1 < time.Now().Unix() {
				enemyDirection = rand.Intn(3)
				enemyDirectionTime = time.Now().Unix()
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
					// 爆発音
					sePlayer := g.audioContext.NewPlayerFromBytes(g.explosionSEBytes)
					sePlayer.Play()
				}

				// 自機の弾と敵の当たり判定（敵と自機の弾の数が一緒なので同じループで実施）
				if isHit(g.my.shot[j].shotx, g.my.shot[j].shoty, g.my.shot[j].drawWidth, g.my.shot[j].drawHeight, g.enemy[i].posx, g.enemy[i].posy, g.enemy[i].drawWidth, g.enemy[i].drawHeight) {
					// 当たってるので敵消失
					g.enemy[i].isHit = true
					g.score += 100
					g.my.shot[j].shotFlag = false
					g.my.shot[j].shotx = g.my.posx
					g.my.shot[j].shoty = g.my.posy
					t := time.Now()
					// TODO
					g.enemy[i].revivalTime = t.Add(time.Second + 12000000)
					// 爆発音
					sePlayer := g.audioContext.NewPlayerFromBytes(g.explosionSEBytes)
					sePlayer.Play()
				}
			}
		} else {
			// エフェクトのカウント
			g.enemy[i].effectCount++
			// 時間経ってたら復活
			if g.enemy[i].revivalTime.Before(time.Now()) {
				g.enemy[i].isHit = false
				g.enemy[i].effectCount = 0
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
			//ebitenutil.DebugPrint(screen, "enemyHit!!!")
			// 爆発表示
			if g.enemy[i].effectCount < 54 {
				op1 := &ebiten.DrawImageOptions{}
				op1.GeoM.Translate(-float64(EFFECT_FRAME_WIDTH)/2, -float64(EFFECT_FRAME_HEIGHT)/2)
				g.enemy[i].drawWidth = float64(g.enemy[i].drawImage.Bounds().Dx()) * 0.3
				g.enemy[i].drawHeight = float64(g.enemy[i].drawImage.Bounds().Dy()) * 0.3
				op1.GeoM.Translate(g.enemy[i].posx+float64(g.enemy[i].drawWidth/2), g.enemy[i].posy+float64(g.enemy[i].drawHeight/2))
				rx := (g.enemy[i].effectCount / 6) % EFFECT_FRAME_COUNT
				sx := EFFECT_FRAMEOX + (rx%3)*EFFECT_FRAME_WIDTH
				sy := EFFECT_FRAMEOY
				if g.enemy[i].effectCount >= 18 {
					sy = EFFECT_FRAMEOX + EFFECT_FRAME_HEIGHT
				}
				if g.enemy[i].effectCount >= 36 {
					sy = EFFECT_FRAMEOY + EFFECT_FRAME_HEIGHT*2
				}
				screen.DrawImage(g.enemy[i].effectImage.SubImage(image.Rect(sx, sy, sx+EFFECT_FRAME_WIDTH, sy+EFFECT_FRAME_HEIGHT)).(*ebiten.Image), op1)
			}
			// 敵非表示
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
	if !g.my.isHit {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(0.1, 0.1)
		op.GeoM.Translate(g.my.posx, g.my.posy)
		screen.DrawImage(g.my.drawImage, op)
		g.my.drawWidth = float64(g.my.drawImage.Bounds().Dx()) * 0.1
		g.my.drawHeight = float64(g.my.drawImage.Bounds().Dy()) * 0.1
	} else {
		// 爆発表示
		if g.my.effectCount < 54 {
			op1 := &ebiten.DrawImageOptions{}
			op1.GeoM.Translate(-float64(EFFECT_FRAME_WIDTH)/2, -float64(EFFECT_FRAME_HEIGHT)/2)
			op1.GeoM.Translate(g.my.posx+float64(g.my.drawWidth/2), g.my.posy+float64(g.my.drawHeight/2))
			rx := (g.my.effectCount / 6) % EFFECT_FRAME_COUNT
			sx := EFFECT_FRAMEOX + (rx%3)*EFFECT_FRAME_WIDTH
			sy := EFFECT_FRAMEOY
			if g.my.effectCount >= 18 {
				sy = EFFECT_FRAMEOX + EFFECT_FRAME_HEIGHT
			}
			if g.my.effectCount >= 36 {
				sy = EFFECT_FRAMEOY + EFFECT_FRAME_HEIGHT*2
			}
			screen.DrawImage(g.my.effectImage.SubImage(image.Rect(sx, sy, sx+EFFECT_FRAME_WIDTH, sy+EFFECT_FRAME_HEIGHT)).(*ebiten.Image), op1)
		}
	}

	/*
		ebitenutil.DebugPrint(screen, fmt.Sprintf("score:%d", g.score))
		if g.my.isHit {
			ebitenutil.DebugPrint(screen, fmt.Sprintf("GameOver!!!:%d", g.score))
		}
	*/
	// スコア表示（桁数分ループしてベース画像から数字を引っ張ってくる）
	for i := 0; i <= 6; i++ {

		s := getDigits(g.score, i, i)

		op4 := &ebiten.DrawImageOptions{}
		op4.GeoM.Scale(0.5, 0.5)
		// スコアの数字から取得位置変える
		// 桁数から表示位置を変える
		//op1.GeoM.Translate(-float64(EFFECT_FRAME_WIDTH)/2, -float64(EFFECT_FRAME_HEIGHT)/2)
		op4.GeoM.Translate(float64(500-(i*32)), 0)
		sx := SCORE_OX + (s * 32)
		sy := SCORE_OY

		screen.DrawImage(g.scoreImage.SubImage(image.Rect(sx, sy, sx+SCORE_WIDTH, sy+SCORE_HEIGHT)).(*ebiten.Image), op4)
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

func openSE(name string) ([]byte, error) {
	file, err := os.Open(name)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	src, err2 := mp3.DecodeWithoutResampling(file)
	if err2 != nil {
		panic(err2)
	}
	return io.ReadAll(src)
}

func getDigits(value int, start int, end int) int {
	var mod_value int
	var result int

	// n桁目以下の桁を取得
	mod_value = value % int(math.Pow(10, float64(end)+1))

	// m桁目以上の桁を取得
	result = mod_value / int(math.Pow(10, float64(start)))

	return result
}
