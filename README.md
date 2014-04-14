convert-rf-readings
===================

Allows conversion to/from &lt;band> referenced and &lt;bandless> referenced readings within HouseMon App (v0.9.0)

**Note**:The processing/storage of readings in HouseMon 0.9.0 is an early reference and may change, but this little utility
combined with use of these upgraded core Gadgets (https://github.com/TheDistractor/flow-ext/tree/master/gadgets/housemon/rf)[here]
will allow you to at least handle multiple rf network inputs from say 868Mhz and 433Mhz simultaneously.


To install (First STOP Housemon/Jeebus if its running):

```bash
    #get flow-ext if you dont already have it
    go get https://github.com/TheDistractor/flow-ext
    #get this little utility
    go get https://github.com/TheDistractor/convert-rf-readings
    go build
    ./convert-rf-readings help
    #now run your conversion with referencing the path to your Housemon data folder - something like:
    ./convert-rf-readings --path '../../jcw/housemon/data
```

Then, once you are happy with conversions add the following to your Housemon imports within main.go

```go

	_ "github.com/TheDistractor/flow-ext/gadgets/housemon/rf/nodemap"  //NodeMap
	_ "github.com/TheDistractor/flow-ext/gadgets/housemon/rf/putreadings"  //PutReadings


```

Then edit setup.coffee (or wherever you keep your circuit config for driverFill) and add it the <band> references:

```json

    { data: "RFb433g5i22,radioBlip,BLIPPer",     to: "nm.Info" }
    { data: "RFb868g5i22,radioBlip,RadioBLIP",     to: "nm.Info" }

```
