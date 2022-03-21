# video-color-palette-generator

Generates color palette of a given video. Using k-means for clustering the color of video frames

## Args

[1] : video filepath

[2] : segment duration in seconds

[3] : Output folder for segmented frames. File will have following format `frame_%d__segment_%d.png`

[4] : Color palette size

[5] : Iteration of k-means. Default: 300

```shell
go run .\main.go .\t1.mp4 10 frames_output 6 300
```






