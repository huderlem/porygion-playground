package main

import (
	"bytes"
	"image/png"
	_ "image/png"
	"strconv"
	"syscall/js"
	"time"

	"github.com/huderlem/porygion"
)

type context struct {
	regionMap porygion.RegionMap
}

func (c *context) renderBaseMap() error {
	imgBase := porygion.RenderBaseRegionMap(c.regionMap)
	buf := new(bytes.Buffer)
	err := png.Encode(buf, imgBase)
	if err != nil {
		return err
	}
	dst := js.Global().Get("Uint8Array").New(len(buf.Bytes()))
	js.CopyBytesToJS(dst, buf.Bytes())
	js.Global().Call("displayImage", dst, "region-map-image-base")
	return nil
}

func (c *context) renderCitiesMap() error {
	imgBase := porygion.RenderRegionMapWithCities(c.regionMap)
	buf := new(bytes.Buffer)
	err := png.Encode(buf, imgBase)
	if err != nil {
		return err
	}
	dst := js.Global().Get("Uint8Array").New(len(buf.Bytes()))
	js.CopyBytesToJS(dst, buf.Bytes())
	js.Global().Call("displayImage", dst, "region-map-image-cities")
	return nil
}

func (c *context) renderFullMap() error {
	imgBase := porygion.RenderFullRegionMap(c.regionMap)
	buf := new(bytes.Buffer)
	err := png.Encode(buf, imgBase)
	if err != nil {
		return err
	}
	dst := js.Global().Get("Uint8Array").New(len(buf.Bytes()))
	js.CopyBytesToJS(dst, buf.Bytes())
	js.Global().Call("displayImage", dst, "region-map-image-full")
	return nil
}

func (c *context) generateBase() interface{} {
	seed := time.Now().UnixNano()
	regionMap := porygion.GenerateBaseRegionMap(seed, 240, 160)
	c.regionMap = regionMap
	c.renderBaseMap()
	return nil
}

func (c *context) generateCities() interface{} {
	if len(c.regionMap.Elevations) == 0 {
		return nil
	}
	numCities, _ := strconv.Atoi(js.Global().Get("document").
		Call("getElementById", "num-cities-input").
		Get("value").String())
	seed := time.Now().UnixNano()
	regionMap := porygion.GenerateRegionMapWithCities(seed, numCities, c.regionMap)
	c.regionMap = regionMap
	c.renderBaseMap()
	c.renderCitiesMap()
	return regionMap
}

func (c *context) generateRoutes() interface{} {
	if len(c.regionMap.Elevations) == 0 {
		return nil
	}
	seed := time.Now().UnixNano()
	regionMap, err := porygion.GenerateRegionMapWithRoutes(seed, c.regionMap)
	if err != nil {
		return porygion.RegionMap{}
	}
	c.regionMap = regionMap
	c.renderBaseMap()
	c.renderCitiesMap()
	c.renderFullMap()

	return regionMap
}

func (c *context) generateFull() interface{} {
	seed := time.Now().UnixNano()
	numCities, _ := strconv.Atoi(js.Global().Get("document").
		Call("getElementById", "num-cities-input").
		Get("value").String())
	regionMap, err := porygion.GenerateRegionMap(seed, 240, 160, numCities)
	if err != nil {
		return porygion.RegionMap{}
	}
	c.regionMap = regionMap
	c.renderBaseMap()
	c.renderCitiesMap()
	c.renderFullMap()

	return regionMap
}

func (c *context) registerFunctions() {
	js.Global().Get("document").
		Call("getElementById", "generate-full-button").
		Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			c.generateFull()
			return nil
		}))
	js.Global().Get("document").
		Call("getElementById", "generate-base-button").
		Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			c.generateBase()
			return nil
		}))
	js.Global().Get("document").
		Call("getElementById", "generate-cities-button").
		Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			c.generateCities()
			return nil
		}))
	js.Global().Get("document").
		Call("getElementById", "generate-routes-button").
		Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			c.generateRoutes()
			return nil
		}))
}

func main() {
	c := make(chan struct{}, 0)
	ctx := context{}
	ctx.registerFunctions()
	<-c
}
