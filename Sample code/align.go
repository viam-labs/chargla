package main

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"math"
	"time"
	
	"github.com/edaniels/golog"
	"go.viam.com/rdk/components/base"
	"go.viam.com/rdk/components/camera"
	"go.viam.com/rdk/rimage"
	"go.viam.com/rdk/robot/client"
	"go.viam.com/rdk/utils"
	"go.viam.com/utils/rpc"

	"github.com/bopoh24/gocv"

    "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"

	"github.com/fogleman/gg"
)

type circle struct {
	location image.Point
	radius int
}

func findChargingCirclesHelper(m gocv.Mat, p1, p2 float64) []circle {
	circles := gocv.NewMat()
	defer circles.Close()
	
	gocv.HoughCirclesWithParams(
		m,
		&circles,
		gocv.HoughGradient,
		1,                     // dp
		float64(m.Rows()/16), // minDist
		p1,                    // param1
		p2,                    // param2
		0,                    // minRadius
		50,                     // maxRadius

	)

	arr := []circle{}

	for i := 0; i < circles.Cols(); i++ {
		v := circles.GetVecfAt(0, i)
		// if circles are found
		if len(v) > 2 {
			x := int(v[0])
			y := int(v[1])
			r := int(v[2])

			arr = append(arr, circle{image.Point{x,y}, r})
		}
	}

	return arr
}

func findChargingCircles(img image.Image) ([]circle, error) {
	m, err := gocv.ImageToMatRGB(img)
	if err != nil {
		return nil, err
	}
	defer m.Close()

	gocv.CvtColor(m, &m, gocv.ColorRGBToGray)

	for x := 0.0; x < 5; x += 1 {
		arr := findChargingCirclesHelper(m, 200, 33 + x)
		if len(arr) == 2 {
			return arr, nil
		}

		arr = findChargingCirclesHelper(m, 200, 33 - x)
		if len(arr) == 2 {
			return arr, nil
		}
	}

	return nil, fmt.Errorf("cannot find 2 charging circles")
}

func main() {
	logger := golog.NewDevelopmentLogger("client")
	err := mainWithErr(logger)
	if err != nil {
		logger.Fatal(err)
	}
}

func mainWithErr(logger golog.Logger) error {
	a := app.New()
    w := a.NewWindow("Images")
	
	ctx := context.Background()
	robot, err := client.New(
		ctx,
		"bot1-main.47b83dmplo.viam.cloud",
		//"pi2.47b83dmplo.viam.cloud",
		logger,
		client.WithDialOptions(rpc.WithCredentials(rpc.Credentials{
			Type:    utils.CredentialsTypeRobotLocationSecret,
			Payload: "1yq41a679kopz4qnzmba58vqcj90zc32454ssmvez9j7ssk7",
		})),
	)
	if err != nil {
		return err
	}

	defer robot.Close(context.Background())
	
	lasercam, err := camera.FromRobot(robot, "chargercam")
	if err != nil {
		return err
	}

	myBase, err := base.FromRobot(robot, "base")
	if err != nil {
		return err
	}
	
	go func() {
		for {
			logger.Info("going to get image")
			img, closer, err := camera.ReadImage(ctx, lasercam)
			if err != nil {
				panic(err)
			}
			defer closer()
			logger.Info("\t got image")

			rimage.WriteImageToFile(fmt.Sprintf("data/img-%d.jpg", time.Now().Unix()), img)
			
			circles, err := findChargingCircles(img)
			if err != nil {
				logger.Error(err)
			} 
			logger.Info(circles)

			if len(circles) == 2 {
				centerRatio := float64((circles[0].location.X + circles[1].location.X) / 2) / float64(img.Bounds().Max.X)
				degrees := 40 * (.5 - centerRatio)
				logger.Infof("\tcenterRatio: %v degrees: %v", centerRatio, degrees)

				if math.Abs(degrees) < 5 {
					err = myBase.MoveStraight(ctx, 30, 100, nil)
				} else {
					logger.Info("\t\t spining")
					err = myBase.Spin(ctx, degrees, 10000, nil)
				}
				if err != nil {
					logger.Error(err)
				}
			}
			
			draw := gg.NewContextForImage(img)

			for _, c := range circles {
				draw.DrawCircle(float64(c.location.X), float64(c.location.Y), float64(c.radius))
				draw.SetColor(&color.RGBA{255, 0, 0, 1})
				draw.SetLineWidth(10)
				draw.Fill()
			}
			i := canvas.NewImageFromImage(draw.Image())
			
			i.FillMode = canvas.ImageFillOriginal
			w.SetContent(i)

			
		}
	}()

	logger.Info("hello")
	w.ShowAndRun()


	return nil
	
}
