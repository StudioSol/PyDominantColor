// Copyright (c) 2011 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package dominantcolor provides a function for finding
// a color that represents the calculated dominant color in the
// image. This uses a KMean clustering algorithm to find clusters of pixel
// colors in RGB space.
//
// The algorithm is ported from Chromium source code:
//     https://src.chromium.org/svn/trunk/src/ui/gfx/color_analysis.h
//     https://src.chromium.org/svn/trunk/src/ui/gfx/color_analysis.cc
//
// RGB KMean Algorithm (N clusters, M iterations):
//
// 1. Pick N starting colors by randomly sampling the pixels. If you see a
// color you already saw keep sampling. After a certain number of tries
// just remove the cluster and continue with N = N-1 clusters (for an image
// with just one color this should devolve to N=1). These colors are the
// centers of your N clusters.
//
// 2. For each pixel in the image find the cluster that it is closest to in RGB
// space. Add that pixel's color to that cluster (we keep a sum and a count
// of all of the pixels added to the space, so just add it to the sum and
// increment count).
//
// 3. Calculate the new cluster centroids by getting the average color of all of
// the pixels in each cluster (dividing the sum by the count).
//
// 4. See if the new centroids are the same as the old centroids.
//
// a) If this is the case for all N clusters than we have converged and can move on.
//
// b) If any centroid moved, repeat step 2 with the new centroids for up to M iterations.
//
// 5. Once the clusters have converged or M iterations have been tried, sort
// the clusters by weight (where weight is the number of pixels that make up
// this cluster).
//
// 6. Going through the sorted list of clusters, pick the first cluster with the
// largest weight that's centroid falls between |lower_bound| and
// |upper_bound|. Return that color.
// If no color fulfills that requirement return the color with the largest
// weight regardless of whether or not it fulfills the equation above.
package dominantcolor

import (
	"image"
	"image/color"
	"math/rand"
	"sort"

	"github.com/nfnt/resize"
)

type DominantColor struct {
	SampleImageSize            uint
	NumberOfClusters           int
	UniqueColorSearchRetries   int
	ConvergenceIterations      int
	MaximumBrightnessThreshold uint16
	MaximumDarknessThreshold   uint16
}

func (d *DominantColor) FromImage(img image.Image) color.RGBA {
	// Shrink image for faster processing.
	img = resize.Thumbnail(d.SampleImageSize, d.SampleImageSize, img, resize.NearestNeighbor)

	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	rnd := rand.New(rand.NewSource(0))
	randomPoint := func() (x, y int) {
		x = bounds.Min.X + rnd.Intn(width)
		y = bounds.Min.Y + rnd.Intn(height)
		return
	}
	// Pick a starting point for each cluster.
	clusters := make(kMeanClusterGroup, 0, d.NumberOfClusters)
	for i := 0; i < d.NumberOfClusters; i++ {
		// Try up to 10 times to find a unique color. If no unique color can be
		// found, destroy this cluster.
		colorUnique := false
		for j := 0; j < d.UniqueColorSearchRetries; j++ {
			ri, gi, bi, a := img.At(randomPoint()).RGBA()
			// Ignore transparent pixels.
			if a == 0 {
				continue
			}
			r, g, b := uint8(ri/255), uint8(gi/255), uint8(bi/255)
			// Check to see if we have seen this color before.
			colorUnique = !clusters.ContainsCentroid(r, g, b)
			// If we have a unique color set the center of the cluster to
			// that color.
			if colorUnique {
				c := new(kMeanCluster)
				c.SetCentroid(r, g, b)
				clusters = append(clusters, c)
				break
			}
		}
		if !colorUnique {
			break
		}
	}
	convergence := false
	for i := 0; i < d.ConvergenceIterations && !convergence && len(clusters) != 0; i++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				ri, gi, bi, a := img.At(x, y).RGBA()
				// Ignore transparent pixels.
				if a == 0 {
					continue
				}
				r, g, b := uint8(ri/255), uint8(gi/255), uint8(bi/255)
				// Figure out which cluster this color is closest to in RGB space.
				closest := clusters.Closest(r, g, b)
				closest.AddPoint(r, g, b)
			}
		}
		// Calculate the new cluster centers and see if we've converged or not.
		convergence = true
		for _, c := range clusters {
			convergence = convergence && c.CompareCentroidWithAggregate()
			c.RecomputeCentroid()
		}
	}
	// Sort the clusters by population so we can tell what the most popular
	// color is.
	sort.Sort(byWeight(clusters))
	// Loop through the clusters to figure out which cluster has an appropriate
	// color. Skip any that are too bright/dark and go in order of weight.
	var col color.RGBA
	for i, c := range clusters {
		r, g, b := c.Centroid()
		// Sum the RGB components to determine if the color is too bright or too dark.
		summedColor := uint16(r) + uint16(g) + uint16(b)

		if summedColor < d.MaximumBrightnessThreshold && summedColor > d.MaximumDarknessThreshold {
			// If we found a valid color just set it and break. We don't want to
			// check the other ones.
			col.R = r
			col.G = g
			col.B = b
			col.A = 0xFF
			break
		} else if i == 0 {
			// We haven't found a valid color, but we are at the first color so
			// set the color anyway to make sure we at least have a value here.
			col.R = r
			col.G = g
			col.B = b
			col.A = 0xFF
		}
	}
	return col
}

// NewDefault creates a new instance of DominantColor with
// default settings
func NewDefault() *DominantColor {
	return &DominantColor{
		SampleImageSize:            256,
		NumberOfClusters:           4,
		UniqueColorSearchRetries:   10,
		ConvergenceIterations:      50,
		MaximumBrightnessThreshold: 665,
		MaximumDarknessThreshold:   100,
	}
}

// New creates a new instance of DominantColor
func New(sampleImageSize uint, numberOfClusters, uniqueColorSearchRetries,
	convergenceIterations int, maximumBrightnessThreshold,
	maximumDarknessThreshold uint16) *DominantColor {

	return &DominantColor{
		SampleImageSize:            sampleImageSize,
		NumberOfClusters:           numberOfClusters,
		UniqueColorSearchRetries:   uniqueColorSearchRetries,
		ConvergenceIterations:      convergenceIterations,
		MaximumBrightnessThreshold: maximumBrightnessThreshold,
		MaximumDarknessThreshold:   maximumDarknessThreshold,
	}
}
