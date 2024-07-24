package parser

import (
	"log"

	"github.com/aichingert/dxf/pkg/blocks"
	"github.com/aichingert/dxf/pkg/entity"
)

// TODO: replace with actual entities values
var (
	coords2D = [2]float64{0, 0}
	coords3D = [3]float64{0, 0, 0}
)

func ParseAcDbEntity(r *Reader, entity entity.Entity) {
	r.ConsumeNumber(5, HexRadix, "handle", entity.GetHandle())

	// TODO: set hard owner/handle to owner dictionary
	if r.ConsumeStrIf(102, nil) { // consumeIf => ex. {ACAD_XDICTIONARY
		r.ConsumeStr(nil) // 360 => hard owner
		for r.ConsumeNumberIf(330, HexRadix, "soft owner", nil) {
		}
		r.ConsumeStr(nil) // 102 }
	}

	if r.ConsumeStrIf(102, nil) { // consumeIf => ex. {ACAD_XDICTIONARY
		r.ConsumeStr(nil) // 360 => hard owner
		for r.ConsumeNumberIf(330, HexRadix, "soft owner", nil) {
		}
		r.ConsumeStr(nil) // 102 }
	}

	r.ConsumeNumber(330, HexRadix, "owner ptr", entity.GetOwner())

	if r.AssertNextLine("AcDbEntity") != nil {
		return
	}

	// TODO: think about paper space visibility
	r.ConsumeStrIf(67, nil)
	r.ConsumeStr(entity.GetLayerName())

	r.ConsumeStrIf(6, nil) // ByBlock
	r.ConsumeNumberIf(62, DecRadix, "color number (present if not bylayer)", nil)
	r.ConsumeFloatIf(48, "linetype scale", nil)
	r.ConsumeNumberIf(60, DecRadix, "object visibility", entity.GetVisibility())

	r.ConsumeNumberIf(420, DecRadix, "24-bit color value", nil)
	r.ConsumeNumberIf(440, DecRadix, "transparency value", nil)
	r.ConsumeNumberIf(370, DecRadix, "not documented", nil)
}

func ParseAcDbLine(r *Reader, line *entity.Line) {
	if r.AssertNextLine("AcDbLine") != nil {
		return
	}

	r.ConsumeFloatIf(39, "thickness", nil)
	r.ConsumeCoordinates(line.Src[:])
	r.ConsumeCoordinates(line.Dst[:])
}

func ParseAcDbPolyline(r *Reader, polyline *entity.Polyline) {
	if r.AssertNextLine("AcDbPolyline") != nil {
		return
	}

	vertices := int64(0)
	r.ConsumeNumber(90, DecRadix, "number of vertices", &vertices)
	r.ConsumeNumber(70, DecRadix, "polyline flag", &polyline.Flag)

	if !r.ConsumeFloatIf(43, "line width for each vertex", nil) {
		//r.ConsumeFloat(43, "", nil)
		//log.Fatal("[ENTITIES(", Line, ")] TODO: implement line width for each vertex")
	}

	for i := int64(0); i < vertices; i++ {
		bulge := 0.0

		r.ConsumeCoordinates(coords2D[:])

		r.ConsumeFloatIf(40, "default start width", nil)
		r.ConsumeFloatIf(41, "default end width", nil)

		r.ConsumeFloatIf(42, "expected bulge", &bulge)
		r.ConsumeNumberIf(91, DecRadix, "vertex identifier", nil)

		if r.Err() != nil {
			return
		}

		polyline.AppendPLine(coords2D, bulge)
	}
}

func ParseAcDb2dPolyline(r *Reader, _ *entity.Polyline) {
	if r.AssertNextLine("AcDb2dPolyline") != nil {
		return
	}

	r.ConsumeNumberIf(66, DecRadix, "obsolete", nil)
	r.ConsumeCoordinates(coords3D[:])
	r.ConsumeFloatIf(39, "thickness", nil)
	r.ConsumeNumberIf(70, DecRadix, "polyline flag", nil)

	r.ConsumeFloatIf(40, "start width default 0", nil)
	r.ConsumeFloatIf(41, "end width default 0", nil)
	r.ConsumeFloatIf(71, "mesh M vertex count", nil)
	r.ConsumeFloatIf(72, "mesh N vertex count", nil)
	r.ConsumeFloatIf(73, "smooth surface M density", nil)
	r.ConsumeFloatIf(74, "smooth surface N density", nil)
	r.ConsumeNumberIf(75, DecRadix, "curves and smooth surface default 0", nil)

	r.ConsumeCoordinatesIf(210, coords3D[:])
}

func ParseAcDbCircle(r *Reader, circle *entity.Circle) {
	if r.AssertNextLine("AcDbCircle") != nil {
		return
	}

	r.ConsumeFloatIf(39, "thickness", nil)
	r.ConsumeCoordinates(circle.Coordinates[:])
	r.ConsumeFloat(40, "expected radius", &circle.Radius)
}

func ParseAcDbArc(r *Reader, arc *entity.Arc) {
	if r.AssertNextLine("AcDbArc") != nil {
		return
	}

	r.ConsumeFloat(50, "expected startAngle", &arc.StartAngle)
	r.ConsumeFloat(51, "expected endAngle", &arc.EndAngle)
}

func ParseAcDbText(r *Reader, text *entity.Text) {
	if r.AssertNextLine("AcDbText") != nil {
		return
	}

	r.ConsumeFloatIf(39, "expected thickness", &text.Thickness)
	r.ConsumeCoordinates(text.Coordinates[:])

	r.ConsumeFloat(40, "expected text height", &text.Height)
	r.ConsumeStr(&text.Text) // [1] default value of the string itself

	r.ConsumeFloatIf(50, "text rotation default 0", &text.Rotation)
	r.ConsumeFloatIf(41, "relative x scale factor default 1", &text.XScale)
	r.ConsumeFloatIf(51, "oblique angle default 0", &text.Oblique)

	r.ConsumeStrIf(7, &text.Style) // text style name default STANDARD

	r.ConsumeNumberIf(71, DecRadix, "text generation flags default 0", &text.Flags)
	r.ConsumeNumberIf(72, DecRadix, "horizontal text justification", &text.HJustification)

	r.ConsumeCoordinatesIf(11, text.Vector[:])
	r.ConsumeCoordinatesIf(210, text.Vector[:])

	line, _ := r.PeekLine()
	if line == "AcDbText" {
		r.ConsumeStr(nil) // second AcDbText (optional)
	}

	// group 72 and 73 integer codes
	// https://help.autodesk.com/view/OARX/2024/ENU/?guid=GUID-62E5383D-8A14-47B4-BFC4-35824CAE8363

	r.ConsumeNumberIf(73, DecRadix, "vertical text justification", &text.VJustification)
}

func ParseAcDbMText(r *Reader, mText *entity.MText) {
	if r.AssertNextLine("AcDbMText") != nil {
		return
	}

	r.ConsumeCoordinates(mText.Coordinates[:])
	r.ConsumeFloat(40, "expected text height", &mText.TextHeight)

	// TODO: https://ezdxf.readthedocs.io/en/stable/dxfinternals/entities/mtext.html
	r.ConsumeFloat(41, "rectangle width", nil)
	r.ConsumeFloat(46, "column height", nil)

	r.ConsumeNumber(71, DecRadix, "attachment point", &mText.Layout)
	r.ConsumeNumber(72, DecRadix, "direction (ex: left to right)", &mText.Direction)

	// TODO: implement more helper :smelting:
	code, err := r.PeekCode()
	if err != nil {
		return
	}

	for code == 1 || code == 3 {
		line := r.ConsumeDxfLine()
		if r.err != nil {
			return
		}

		mText.Text = append(mText.Text, line.Line)

		code, err = r.PeekCode()
		if err != nil {
			return
		}
	}

	r.ConsumeStrIf(7, &mText.TextStyle)
	r.ConsumeCoordinatesIf(11, mText.Vector[:])

	r.ConsumeNumber(73, DecRadix, "line spacing", &mText.LineSpacing)
	r.ConsumeFloat(44, "line spacing factor", nil)
	HelperParseEmbeddedObject(r)
}

func HelperParseEmbeddedObject(r *Reader) {
	// Embedded Object
	if r.ConsumeStrIf(101, nil) {
		r.ConsumeNumberIf(70, DecRadix, "not documented", nil)
		r.ConsumeCoordinates(coords3D[:])
		r.ConsumeCoordinatesIf(11, coords3D[:])

		r.ConsumeFloatIf(40, "not documented", nil)
		r.ConsumeFloatIf(41, "not documented", nil)
		r.ConsumeFloatIf(42, "not documented", nil)
		r.ConsumeFloatIf(43, "not documented", nil)
		r.ConsumeFloatIf(46, "not documented", nil)

		r.ConsumeNumberIf(71, DecRadix, "not documented", nil)
		r.ConsumeNumberIf(72, DecRadix, "not documented", nil)
		r.ConsumeStrIf(1, nil)

		r.ConsumeFloatIf(44, "not documented", nil)
		r.ConsumeFloatIf(45, "not documented", nil)

		r.ConsumeNumberIf(73, DecRadix, "not documented", nil)
		r.ConsumeNumberIf(74, DecRadix, "not documented", nil)

		r.ConsumeFloatIf(44, "not documented", nil)
		r.ConsumeFloatIf(46, "not documented", nil)
	}
}

func ParseAcDbHatch(r *Reader, hatch *entity.Hatch) {
	if r.AssertNextLine("AcDbHatch") != nil {
		return
	}

	r.ConsumeCoordinates(coords3D[:]) // elevation
	r.ConsumeCoordinates(coords3D[:])

	r.ConsumeStr(&hatch.PatternName)
	r.ConsumeNumber(70, DecRadix, "solid fill flag", &hatch.SolidFill)
	r.ConsumeNumber(71, DecRadix, "associativity flag", &hatch.Associative)

	boundaryPaths := int64(0)
	r.ConsumeNumber(91, DecRadix, "boundary paths", &boundaryPaths)
	for i := int64(0); i < boundaryPaths; i++ {
		hatch.BoundaryPaths = append(hatch.BoundaryPaths, ParseBoundaryPath(r))
	}

	r.ConsumeNumber(75, DecRadix, "hatch style", &hatch.Style)
	r.ConsumeNumber(76, DecRadix, "hatch pattern type", &hatch.Pattern)
	r.ConsumeFloatIf(52, "hatch pattern angle", &hatch.Angle)
	r.ConsumeFloatIf(41, "hatch pattern scale or spacing", &hatch.Scale)
	r.ConsumeNumberIf(77, DecRadix, "hatch pattern double flag", &hatch.Double)

	patternDefinitions := int64(0)

	r.ConsumeNumberIf(78, DecRadix, "number of pattern definition lines", &patternDefinitions)

	for i := int64(0); i < patternDefinitions; i++ {
		base, offset, angle := [2]float64{0.0, 0.0}, [2]float64{0.0, 0.0}, 0.0
		var dashes []float64
		dashLen := 0.0

		r.ConsumeFloat(53, "pattern line angle", &angle)
		r.ConsumeFloat(43, "pattern line base point x", &base[0])
		r.ConsumeFloat(44, "pattern line base point y", &base[1])
		r.ConsumeFloat(45, "pattern line offset x", &offset[0])
		r.ConsumeFloat(46, "pattern line offset y", &offset[1])

		dashLengths := int64(0)
		r.ConsumeNumber(79, DecRadix, "number of dash length items", &dashLengths)

		for j := int64(0); j < dashLengths; j++ {
			r.ConsumeFloat(49, "dash length", &dashLen)
			dashes = append(dashes, dashLen)
		}

		hatch.AppendPatternLine(angle, base, offset, dashes)
	}

	r.ConsumeFloatIf(47, "pixel size used to determine the density", &hatch.PixelSize)

	seedPoints, nColors := int64(0), int64(0)
	r.ConsumeNumber(98, DecRadix, "number of seed points", &seedPoints)

	for seedPoint := int64(0); seedPoint < seedPoints; seedPoint++ {
		r.ConsumeCoordinates(hatch.SeedPoint[:2])
	}

	r.ConsumeNumberIf(450, DecRadix, "indicates solid hatch or gradient", nil)
	r.ConsumeNumberIf(451, DecRadix, "zero is reserved for future use", nil)
	r.ConsumeFloatIf(460, "rotation angle in radians for gradients", nil)
	r.ConsumeFloatIf(461, "gradient definition", nil)
	r.ConsumeNumberIf(452, DecRadix, "records how colors were defined", nil)
	r.ConsumeFloatIf(462, "color tint value used by dialog", nil)
	r.ConsumeNumberIf(453, DecRadix, "number of colors", &nColors)
	for color := int64(0); color < nColors; color++ {
		r.ConsumeFloatIf(463, "reserved for future use", nil)
		r.ConsumeNumberIf(63, DecRadix, "not documented", nil)
		r.ConsumeNumberIf(421, DecRadix, "not documented", nil)
	}
	r.ConsumeStrIf(470, nil) // string default = LINEAR
}

func ParseBoundaryPath(r *Reader) *entity.BoundaryPath {
	path := &entity.BoundaryPath{}

	// [92] Boundary path type flag (bit coded):
	// 0 = Default | 1 = External | 2  = Polyline
	// 4 = Derived | 8 = Textbox  | 16 = Outermost
	r.ConsumeNumber(92, DecRadix, "boundary path type flag", &path.Flag)

	if path.Flag&2 == 2 {
		path.Polyline = &entity.Polyline{}
		hasBulge, bulge, vertices := int64(0), 0.0, int64(0)

		r.ConsumeNumber(72, DecRadix, "has bulge flag", &hasBulge)
		r.ConsumeNumber(73, DecRadix, "is closed flag", &path.Polyline.Flag)
		r.ConsumeNumber(93, DecRadix, "number of polyline vertices", &vertices)

		for vertex := int64(0); vertex < vertices; vertex++ {
			r.ConsumeCoordinates(coords2D[:])
			if hasBulge == 1 {
				r.ConsumeFloat(42, "expected bulge", &bulge)
			}
			path.Polyline.AppendPLine(coords2D, bulge)
		}
	} else {
		edges, edgeType := int64(0), int64(0)

		r.ConsumeNumber(93, DecRadix, "number of edges in this boundary path", &edges)

		for edge := int64(0); edge < edges; edge++ {
			r.ConsumeNumber(72, DecRadix, "edge type data", &edgeType)

			switch edgeType {
			case 1: // Line
				line := entity.NewLine()
				line.Entity = nil
				r.ConsumeCoordinates(line.Src[:2])
				r.ConsumeCoordinates(line.Dst[:2])
				path.Lines = append(path.Lines, line)
			case 2: // Circular arc
				arc := entity.NewArc()
				arc.Entity = nil
				r.ConsumeCoordinates(arc.Circle.Coordinates[:2])
				r.ConsumeFloat(40, "radius", &arc.Circle.Radius)

				r.ConsumeFloat(50, "start angle", &arc.StartAngle)
				r.ConsumeFloat(51, "end angle", &arc.EndAngle)
				r.ConsumeNumber(73, DecRadix, "is counterclockwise", &arc.Counterclockwise)
				path.Arcs = append(path.Arcs, arc)
			case 3: // Elliptic arc
				ellipse := entity.NewEllipse()
				ellipse.Entity = nil

				r.ConsumeCoordinates(ellipse.Center[:2])
				r.ConsumeCoordinates(ellipse.EndPoint[:2])
				r.ConsumeFloat(40, "length of minor axis", &ellipse.Ratio)
				r.ConsumeFloat(50, "start angle", &ellipse.Start)
				r.ConsumeFloat(51, "end angle", &ellipse.End)
				r.ConsumeFloat(73, "is counterclockwise", nil)

				path.Ellipses = append(path.Ellipses, ellipse)
			case 4: // Spine
				log.Fatal("[AcDbHatch(", Line, ")] TODO: implement boundary path spline")
			default:
				log.Println("[AcDbHatch(", Line, ")] invalid edge type data", edgeType)
				r.err = NewParseError("invalid edge type data")
				return path
			}
		}
	}

	boundaryObjectSize, boundaryObjectRef := int64(0), int64(0)
	r.ConsumeNumber(97, DecRadix, "number of source boundary objects", &boundaryObjectSize)
	for i := int64(0); i < boundaryObjectSize; i++ {
		r.ConsumeNumber(330, HexRadix, "reference to source object", &boundaryObjectRef)
	}

	return path
}

func ParseAcDbEllipse(r *Reader, ellipse *entity.Ellipse) {
	if r.AssertNextLine("AcDbEllipse") != nil {
		return
	}

	r.ConsumeCoordinates(ellipse.Center[:])   // Center point
	r.ConsumeCoordinates(ellipse.EndPoint[:]) // Endpoint of major axis

	r.ConsumeCoordinatesIf(210, coords3D[:])

	r.ConsumeFloat(40, "ratio of minor axis to major axis", &ellipse.Ratio)
	r.ConsumeFloat(41, "start parameter", &ellipse.Start)
	r.ConsumeFloat(42, "end parameter", &ellipse.End)
}

func ParseAcDbSpline(r *Reader, _ *entity.MText) {
	if r.AssertNextLine("AcDbSpline") != nil {
		return
	}

	knots, controlPoints, fitPoints := int64(0), int64(0), int64(0)

	r.ConsumeCoordinates(coords3D[:])
	r.ConsumeNumber(70, DecRadix, "spline flag", nil)
	r.ConsumeNumber(71, DecRadix, "degree of the spline curve", nil)
	r.ConsumeNumber(72, DecRadix, "number of knots", &knots)
	r.ConsumeNumber(73, DecRadix, "number of control points", &controlPoints)
	r.ConsumeNumber(74, DecRadix, "number of fit points", &fitPoints)
	r.ConsumeFloatIf(42, "knot tolerance default 0.0000001", nil)
	r.ConsumeFloatIf(43, "control point tolerance 0.0000001", nil)
	r.ConsumeFloatIf(44, "fit tolerance default 0.0000001", nil)

	for i := int64(0); i < knots; i++ {
		r.ConsumeFloat(40, "knot value", nil)
	}
	for i := int64(0); i < controlPoints; i++ {
		r.ConsumeCoordinates(coords3D[:]) // start tangent - may be omitted
	}
	for i := int64(0); i < fitPoints; i++ {
		r.ConsumeCoordinates(coords3D[:]) // end tangent   - may be omitted
	}
}

// AcDbPoint
func ParseAcDbTrace(r *Reader, _ *entity.MText) {
	if r.AssertNextLine("AcDbTrace") != nil {
		return
	}

	r.ConsumeCoordinates(coords3D[:])
	r.ConsumeCoordinates(coords3D[:])
	r.ConsumeCoordinates(coords3D[:])
	r.ConsumeCoordinates(coords3D[:])

	r.ConsumeNumberIf(39, DecRadix, "thickness", nil)
	r.ConsumeCoordinatesIf(210, coords3D[:])
	r.ConsumeFloatIf(50, "angle of the x axis", nil)
}

// TODO: implement entity entity.Vertex
func ParseAcDbVertex(r *Reader, _ *entity.MText) {
	if r.AssertNextLine("AcDbVertex") != nil {
		return
	}

	next := ""
	r.ConsumeStr(&next) // AcDb2dVertex or AcDb3dPolylineVertex

	r.ConsumeCoordinates(coords3D[:])
	r.ConsumeFloatIf(40, "starting width", nil)
	r.ConsumeFloatIf(41, "end width", nil)
	r.ConsumeFloatIf(42, "bulge", nil)

	r.ConsumeNumberIf(70, DecRadix, "vertex flags", nil)
	r.ConsumeFloatIf(50, "curve fit tangent direction", nil)

	r.ConsumeFloatIf(71, "polyface mesh vertex index", nil)
	r.ConsumeFloatIf(72, "polyface mesh vertex index", nil)
	r.ConsumeFloatIf(73, "polyface mesh vertex index", nil)
	r.ConsumeFloatIf(74, "polyface mesh vertex index", nil)

	r.ConsumeNumberIf(91, DecRadix, "vertex identifier", nil)
}

// TODO: implement entity entity.Point
func ParseAcDbPoint(r *Reader, _ *entity.MText) {
	if r.AssertNextLine("AcDbPoint") != nil {
		return
	}

	r.ConsumeCoordinates(coords3D[:])
	r.ConsumeNumberIf(39, DecRadix, "thickness", nil)

	// XYZ extrusion direction
	// optional default 0, 0, 1
	r.ConsumeCoordinatesIf(210, coords3D[:])
	r.ConsumeFloatIf(50, "angle of the x axis", nil)
}

func ParseAcDbBlockReference(r *Reader, insert *entity.Insert) {
	line := ""
	r.ConsumeStr(&line)
	if r.Err() != nil || !(line == "AcDbBlockReference" || line == "AcDbMInsertBlock") {
		return
	}

	r.ConsumeNumberIf(66, DecRadix, "attributes follow", &insert.AttributesFollow)
	r.ConsumeStr(&insert.BlockName)
	r.ConsumeCoordinates(insert.Coordinates[:])

	r.ConsumeFloatIf(41, "x scale factor", &insert.Scale[0])
	r.ConsumeFloatIf(42, "y scale factor", &insert.Scale[1])
	r.ConsumeFloatIf(43, "z scale factor", &insert.Scale[2])

	r.ConsumeFloatIf(50, "rotation angle", &insert.Rotation)
	r.ConsumeNumberIf(70, DecRadix, "column count", &insert.ColCount)
	r.ConsumeNumberIf(71, DecRadix, "row count", &insert.RowCount)

	r.ConsumeFloatIf(44, "column spacing", &insert.ColSpacing)
	r.ConsumeFloatIf(45, "row spacing", &insert.RowSpacing)

	// optional default = 0, 0, 1
	// XYZ extrusion direction
	r.ConsumeCoordinatesIf(210, coords3D[:])
}

func ParseAcDbBlockBegin(r *Reader, block *blocks.Block) {
	if r.AssertNextLine("AcDbBlockBegin") != nil {
		return
	}

	r.ConsumeStr(&block.BlockName) // [2] block name
	r.ConsumeNumber(70, DecRadix, "block-type flag", &block.Flag)
	r.ConsumeCoordinates(block.Coordinates[:])

	r.ConsumeStr(&block.OtherName) // [3] block name
	r.ConsumeStr(&block.XRefPath)  // [1] Xref path name
}

func ParseAcDbAttribute(r *Reader, attrib *entity.Attrib) {
	if r.AssertNextLine("AcDbAttribute") != nil {
		return
	}

	r.ConsumeStr(&attrib.Tag) // [2] Attribute tag
	r.ConsumeNumber(70, DecRadix, "attribute flags", &attrib.Flags)
	r.ConsumeNumberIf(74, DecRadix, "vertical text justification", &attrib.Text.VJustification) // group code 73 TEXT
	r.ConsumeNumberIf(280, DecRadix, "version number", nil)

	r.ConsumeNumberIf(73, DecRadix, "field length", nil) // not currently used
	r.ConsumeFloatIf(50, "text rotation", &attrib.Text.Rotation)
	r.ConsumeFloatIf(41, "relative x scale factor (width)", &attrib.Text.XScale) // adjusted when fit-type text is used
	r.ConsumeFloatIf(51, "oblique angle", &attrib.Text.Oblique)
	r.ConsumeStrIf(7, &attrib.Text.Style) // text style name default STANDARD
	r.ConsumeNumberIf(71, DecRadix, "text generation flags", &attrib.Text.Flags)
	r.ConsumeNumberIf(72, DecRadix, "horizontal text justification", &attrib.Text.HJustification)

	r.ConsumeCoordinatesIf(11, attrib.Text.Vector[:])
	r.ConsumeCoordinatesIf(210, attrib.Text.Vector[:])

	// TODO: parse XDATA
	code, err := r.PeekCode()
	for code != 0 && err == nil {
		r.ConsumeStr(nil)
		code, err = r.PeekCode()
	}
}

func ParseAcDbAttributeDefinition(r *Reader, attdef *entity.Attdef) {
	if r.AssertNextLine("AcDbAttributeDefinition") != nil {
		return
	}

	r.ConsumeStr(&attdef.Prompt) // [3] prompt string
	r.ConsumeStr(&attdef.Tag)    // [2] tag string
	r.ConsumeNumber(70, DecRadix, "attribute flags", &attdef.Flags)
	r.ConsumeFloatIf(73, "field length", nil)
	r.ConsumeNumberIf(74, DecRadix, "vertical text justification", &attdef.Text.VJustification)

	r.ConsumeNumber(280, DecRadix, "lock position flag", nil)

	r.ConsumeNumberIf(71, DecRadix, "attachment point", &attdef.AttachmentPoint)
	r.ConsumeNumberIf(72, DecRadix, "drawing direction", &attdef.DrawingDirection)

	r.ConsumeCoordinatesIf(11, attdef.Direction[:])
	HelperParseEmbeddedObject(r)
}

func ParseAcDbDimension(r *Reader, _ *entity.Attdef) {
	if r.AssertNextLine("AcDbDimension") != nil {
		return
	}

	r.ConsumeNumber(280, DecRadix, "version number", nil)
	r.ConsumeStr(nil) // name of the block

	r.ConsumeCoordinates(coords3D[:])
	r.ConsumeCoordinates(coords3D[:])
	r.ConsumeCoordinatesIf(12, coords3D[:])

	r.ConsumeNumberIf(70, DecRadix, "dimension type", nil)
	r.ConsumeNumberIf(71, DecRadix, "attachment point", nil)
	r.ConsumeNumberIf(72, DecRadix, "dimension text-line spacing", nil)

	r.ConsumeNumberIf(41, DecRadix, "dimension text-line factor", nil)
	r.ConsumeNumberIf(42, DecRadix, "actual measurement", nil)

	r.ConsumeNumberIf(73, DecRadix, "not documented", nil)
	r.ConsumeNumberIf(74, DecRadix, "not documented", nil)
	r.ConsumeNumberIf(75, DecRadix, "not documented", nil)

	r.ConsumeStrIf(1, nil) // dimension text
	r.ConsumeFloatIf(53, "roation angle of the dimension", nil)
	r.ConsumeFloatIf(51, "horizontal direction", nil)

	r.ConsumeNumberIf(71, DecRadix, "attachment point", nil)
	r.ConsumeNumberIf(42, DecRadix, "actual measurement", nil)
	r.ConsumeNumberIf(73, DecRadix, "not documented", nil)
	r.ConsumeNumberIf(74, DecRadix, "not documented", nil)
	r.ConsumeNumberIf(75, DecRadix, "not documented", nil)

	r.ConsumeCoordinatesIf(210, coords3D[:])
	r.ConsumeStrIf(3, nil) // [3] dimension style name

	dim := ""
	r.ConsumeStr(&dim)

	switch dim {
	// should be acdb3pointangulardimension
	case "AcDb2LineAngularDimension":
		r.ConsumeCoordinates(coords3D[:]) // point for linear and angular dimension
		r.ConsumeCoordinates(coords3D[:]) // point for linear and angular dimension
		r.ConsumeCoordinates(coords3D[:]) // point for diameter, radius, and angular dimension
	case "AcDbAlignedDimension":
		r.ConsumeCoordinatesIf(12, coords3D[:]) // insertion point for clones of a dimension
		r.ConsumeCoordinates(coords3D[:])       // definition point for linear and angular dimensions
		r.ConsumeCoordinates(coords3D[:])       // definition point for linear and angular dimensions
		r.ConsumeFloatIf(50, "angle of rotated, horizontal, or vertical dimensions", nil)
		r.ConsumeFloatIf(52, "oblique angle", nil)
		r.ConsumeStrIf(100, nil) // subclass marker AcDbRotatedDimension
	default:
		log.Fatal("Dimension(", Line, ")", dim)
	}
}

func ParseAcDbViewport(r *Reader, _ *entity.MText) {
	if r.AssertNextLine("AcDbViewport") != nil {
		return
	}

	r.ConsumeCoordinates(coords3D[:])
	r.ConsumeFloat(40, "width in paper space units", nil)
	r.ConsumeFloat(41, "height in paper space units", nil)

	// => -1 0 On, 0 = Off
	r.ConsumeFloatIf(68, "viewport status field", nil)

	r.ConsumeNumber(69, DecRadix, "viewport id", nil)

	r.ConsumeCoordinates(coords2D[:]) // center point
	r.ConsumeCoordinates(coords2D[:]) // snap base point
	r.ConsumeCoordinates(coords2D[:]) // snap spacing point
	r.ConsumeCoordinates(coords2D[:]) // grid spacing point
	r.ConsumeCoordinates(coords3D[:]) // view direction vector
	r.ConsumeCoordinates(coords3D[:]) // view target point

	r.ConsumeFloat(42, "perspective lens length", nil)
	r.ConsumeFloat(43, "front clip plane z value", nil)
	r.ConsumeFloat(44, "back clip plane z value", nil)
	r.ConsumeFloat(45, "view height", nil)
	r.ConsumeFloat(50, "snap angle", nil)
	r.ConsumeFloat(51, "view twist angle", nil)

	r.ConsumeFloat(72, "circle zoom percent", nil)

	code, err := r.PeekCode()
	for err != nil && code == 331 {
		r.ConsumeNumber(331, DecRadix, "frozen layer object Id/handle", nil)
	}

	r.ConsumeNumber(90, HexRadix, "viewport status bit-coded flags", nil)
	r.ConsumeNumberIf(340, DecRadix, "hard-pointer id/handle to entity that serves as the viewports clipping boundary", nil)
	r.ConsumeStr(nil) // [1]
	r.ConsumeNumber(281, DecRadix, "render mode", nil)
	r.ConsumeNumber(71, DecRadix, "ucs per viewport flag", nil)
	r.ConsumeNumber(74, DecRadix, "display ucs icon at ucs origin flag", nil)

	r.ConsumeCoordinates(coords3D[:]) // ucs origin
	r.ConsumeCoordinates(coords3D[:]) // ucs x-axis
	r.ConsumeCoordinates(coords3D[:]) // ucs y-axis

	r.ConsumeNumberIf(345, DecRadix, "id/handle of AcDbUCSTableRecord if UCS is a named ucs", nil)
	r.ConsumeNumberIf(346, DecRadix, "id/handle of AcDbUCSTableRecord of base ucs", nil)
	r.ConsumeNumber(79, DecRadix, "Orthographic type of UCS", nil)
	r.ConsumeFloat(146, "elevation", nil)
	r.ConsumeNumber(170, DecRadix, "ShadePlot mode", nil)
	r.ConsumeNumber(61, DecRadix, "frequency of major grid lines compared to minor grid lines", nil)

	r.ConsumeNumberIf(332, DecRadix, "background id/handle", nil)
	r.ConsumeNumberIf(333, DecRadix, "shade plot id/handle", nil)
	r.ConsumeNumberIf(348, DecRadix, "visual style id/handle", nil)

	r.ConsumeNumber(292, DecRadix, "default lighting type on when no use lights are specified", nil)
	r.ConsumeNumber(282, DecRadix, "default lighting type", nil)
	r.ConsumeFloat(141, "view brightness", nil)
	r.ConsumeFloat(142, "view contrast", nil)

	r.ConsumeFloatIf(63, "ambient light color only if not black", nil)
	r.ConsumeFloatIf(421, "ambient light color only if not black", nil)
	r.ConsumeFloatIf(431, "ambient light color only if not black", nil)

	r.ConsumeNumberIf(361, DecRadix, "sun id/handle", nil)
	r.ConsumeNumberIf(335, DecRadix, "soft pointer reference to id/handle", nil)
	r.ConsumeNumberIf(343, DecRadix, "soft pointer reference to id/handle", nil)
	r.ConsumeNumberIf(344, DecRadix, "soft pointer reference to id/handle", nil)
	r.ConsumeNumberIf(91, DecRadix, "soft pointer reference to id/handle", nil)
}
