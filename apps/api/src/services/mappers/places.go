package mappers

import (
	v1 "api/src/generated/google_maps/v1"
	"time"

	"googlemaps.github.io/maps"
)

// --- ToGo Mappers (v1 -> maps) ---

func MapFindPlaceFromTextRequestToGo(req *v1.FindPlaceFromTextRequest) *maps.FindPlaceFromTextRequest {
	if req == nil {
		return nil
	}
	fields := make([]maps.PlaceSearchFieldMask, 0, len(req.Fields))
	for _, f := range req.Fields {
		if str, ok := placeSearchFieldMaskToString[f]; ok {
			fields = append(fields, maps.PlaceSearchFieldMask(str))
		}
	}

	return &maps.FindPlaceFromTextRequest{
		Input:     req.Input,
		InputType: maps.FindPlaceFromTextInputType(inputTypeToString[req.InputType]),
		Language:  req.Language,
		Fields:    fields,
	}
}

func MapFindPlaceFromTextResponseToGo(res *v1.FindPlaceFromTextResponse) *maps.FindPlaceFromTextResponse {
	if res == nil {
		return nil
	}
	candidates := make([]maps.PlacesSearchResult, 0, len(res.Candidates))
	for _, candidate := range res.Candidates {
		if candidate != nil {
			mappedCandidate := mapPlacesSearchResultToGo(candidate)
			candidates = append(candidates, mappedCandidate)
		}
	}

	return &maps.FindPlaceFromTextResponse{
		Candidates:       candidates,
		HTMLAttributions: res.HtmlAttributions,
	}
}

func mapPlacesSearchResultToGo(c *v1.PlacesSearchResult) maps.PlacesSearchResult {
	return maps.PlacesSearchResult{
		FormattedAddress:  c.FormattedAddress,
		Geometry:          mapAddressGeometryToGo(c.Geometry),
		Name:              c.Name,
		Icon:              c.Icon,
		PlaceID:           c.PlaceId,
		Rating:            c.Rating,
		UserRatingsTotal:  int(c.UsersRatingTotal),
		Types:             c.Types,
		OpeningHours:      mapOpeningHoursToGo(c.OpeningHours),
		Photos:            mapPhotosToGo(c.Photos),
		Vicinity:          c.Vicinity,
		PermanentlyClosed: c.PermanentlyClosed,
		BusinessStatus:    c.BusinessStatus,
		PriceLevel:        int(c.PriceLevel),
	}
}

func mapPhotosToGo(ps []*v1.Photo) []maps.Photo {
	if ps == nil {
		return nil
	}
	photos := make([]maps.Photo, 0, len(ps))
	for _, p := range ps {
		if p != nil {
			photos = append(photos, mapPhotoToGo(p))
		}
	}
	return photos
}

func mapPhotoToGo(p *v1.Photo) maps.Photo {
	return maps.Photo{
		PhotoReference:   p.PhotoReference,
		Height:           int(p.Height),
		Width:            int(p.Width),
		HTMLAttributions: p.HtmlAttributions,
	}
}

func mapOpeningHoursToGo(o *v1.OpeningHours) *maps.OpeningHours {
	if o == nil {
		return nil
	}
	return &maps.OpeningHours{
		OpenNow:           &o.OpenNow,
		Periods:           mapOpeningHoursPeriodsToGo(o.Periods),
		WeekdayText:       o.WeekdayText,
		PermanentlyClosed: &o.PermanentlyClosed,
	}
}

func mapOpeningHoursPeriodsToGo(ps []*v1.OpeningHoursPeriod) []maps.OpeningHoursPeriod {
	if ps == nil {
		return nil
	}
	periods := make([]maps.OpeningHoursPeriod, 0, len(ps))
	for _, p := range ps {
		if p != nil {
			periods = append(periods, maps.OpeningHoursPeriod{
				Open:  mapOpeningHoursOpenCloseToGo(p.Open),
				Close: mapOpeningHoursOpenCloseToGo(p.Close),
			})
		}
	}
	return periods
}

func mapOpeningHoursOpenCloseToGo(t *v1.OpeningHoursOpenClose) maps.OpeningHoursOpenClose {
	if t == nil {
		return maps.OpeningHoursOpenClose{}
	}
	return maps.OpeningHoursOpenClose{
		Day:  weekdayMap[t.Day],
		Time: t.Time,
	}
}

func mapAddressGeometryToGo(g *v1.AddressGeometry) maps.AddressGeometry {
	if g == nil {
		return maps.AddressGeometry{}
	}
	return maps.AddressGeometry{
		Location:     mapLatLngToGo(g.Location),
		LocationType: g.LocationType,
		Bounds:       mapLatLngBoundsToGo(g.Bounds),
		Viewport:     mapLatLngBoundsToGo(g.Viewport),
		Types:        g.Types,
	}
}

func mapLatLngToGo(ll *v1.LatLng) maps.LatLng {
	if ll == nil {
		return maps.LatLng{}
	}
	return maps.LatLng{
		Lat: float64(ll.Lat),
		Lng: float64(ll.Lng),
	}
}

func mapLatLngBoundsToGo(llb *v1.LatLngBounds) maps.LatLngBounds {
	if llb == nil {
		return maps.LatLngBounds{}
	}
	return maps.LatLngBounds{
		NorthEast: mapLatLngToGo(llb.NorthEast),
		SouthWest: mapLatLngToGo(llb.SouthWest),
	}
}

// --- ToProto Mappers (maps -> v1) ---

func MapFindPlaceFromTextResponseToProto(res *maps.FindPlaceFromTextResponse) *v1.FindPlaceFromTextResponse {
	if res == nil {
		return nil
	}
	candidates := make([]*v1.PlacesSearchResult, 0, len(res.Candidates))
	for _, candidate := range res.Candidates {
		mappedCandidate := mapPlacesSearchResultToProto(&candidate)
		candidates = append(candidates, mappedCandidate)
	}

	return &v1.FindPlaceFromTextResponse{
		Candidates:       candidates,
		HtmlAttributions: res.HTMLAttributions,
	}
}

func mapPlacesSearchResultToProto(c *maps.PlacesSearchResult) *v1.PlacesSearchResult {
	if c == nil {
		return nil
	}
	
	
	
	return &v1.PlacesSearchResult{
		FormattedAddress:  c.FormattedAddress,
		Geometry:          mapAddressGeometryToProto(&c.Geometry),
		Name:              c.Name,
		Icon:              c.Icon,
		PlaceId:           c.PlaceID,
		Rating:            c.Rating,
		UsersRatingTotal:  int32(c.UserRatingsTotal),
		Types:             c.Types,
		OpeningHours:      mapOpeningHoursToProto(c.OpeningHours),
		Photos:            mapPhotosToProto(c.Photos),
		Vicinity:          c.Vicinity,
		PermanentlyClosed: c.PermanentlyClosed,
		BusinessStatus:    c.BusinessStatus,
		PriceLevel:        int32(c.PriceLevel),
	}
}

func mapPhotosToProto(ps []maps.Photo) []*v1.Photo {
	if ps == nil {
		return nil
	}
	photos := make([]*v1.Photo, 0, len(ps))
	for _, p := range ps {
		photos = append(photos, mapPhotoToProto(&p))
	}
	return photos
}

func mapPhotoToProto(p *maps.Photo) *v1.Photo {
	if p == nil {
		return nil
	}
	return &v1.Photo{
		PhotoReference:   p.PhotoReference,
		Height:           int32(p.Height),
		Width:            int32(p.Width),
		HtmlAttributions: p.HTMLAttributions,
	}
}

func mapOpeningHoursToProto(o *maps.OpeningHours) *v1.OpeningHours {
	if o == nil {
		return nil
	}
	openNow := false
	if o.OpenNow != nil {
		openNow = *o.OpenNow
	}
	permanentlyClosed := false
	if o.PermanentlyClosed != nil {
		permanentlyClosed = *o.PermanentlyClosed
	}

	return &v1.OpeningHours{
		OpenNow:           openNow,
		Periods:           mapOpeningHoursPeriodsToProto(o.Periods),
		WeekdayText:       o.WeekdayText,
		PermanentlyClosed: permanentlyClosed,
	}
}

func mapOpeningHoursPeriodsToProto(ps []maps.OpeningHoursPeriod) []*v1.OpeningHoursPeriod {
	if ps == nil {
		return nil
	}
	periods := make([]*v1.OpeningHoursPeriod, 0, len(ps))
	for _, p := range ps {
		periods = append(periods, &v1.OpeningHoursPeriod{
			Open:  mapOpeningHoursOpenCloseToProto(&p.Open),
			Close: mapOpeningHoursOpenCloseToProto(&p.Close),
		})
	}
	return periods
}

func mapOpeningHoursOpenCloseToProto(t *maps.OpeningHoursOpenClose) *v1.OpeningHoursOpenClose {
	if t == nil {
		return nil
	}
	return &v1.OpeningHoursOpenClose{
		Day:  protoWeekdayMap[t.Day],
		Time: t.Time,
	}
}

func mapAddressGeometryToProto(g *maps.AddressGeometry) *v1.AddressGeometry {
	if g == nil || g.Location.Lat == 0 || g.Location.Lng == 0 || g.LocationType == ""{
		return nil
	}
	return &v1.AddressGeometry{
		Location:     mapLatLngToProto(&g.Location),
		LocationType: g.LocationType,
		Bounds:       mapLatLngBoundsToProto(&g.Bounds),
		Viewport:     mapLatLngBoundsToProto(&g.Viewport),
		Types:        g.Types,
	}
}

func mapLatLngToProto(ll *maps.LatLng) *v1.LatLng {
	if ll == nil {
		return nil
	}
	return &v1.LatLng{
		Lat: float32(ll.Lat),
		Lng: float32(ll.Lng),
	}
}

func mapLatLngBoundsToProto(llb *maps.LatLngBounds) *v1.LatLngBounds {
	if llb == nil {
		return nil
	}
	return &v1.LatLngBounds{
		NorthEast: mapLatLngToProto(&llb.NorthEast),
		SouthWest: mapLatLngToProto(&llb.SouthWest),
	}
}

// --- Conversion Maps ---

var weekdayMap = map[v1.Weekday]time.Weekday{
	v1.Weekday_WEEKDAY_SUNDAY:    time.Sunday,
	v1.Weekday_WEEKDAY_MONDAY:    time.Monday,
	v1.Weekday_WEEKDAY_TUESDAY:   time.Tuesday,
	v1.Weekday_WEEKDAY_WEDNESDAY: time.Wednesday,
	v1.Weekday_WEEKDAY_THURSDAY:  time.Thursday,
	v1.Weekday_WEEKDAY_FRIDAY:    time.Friday,
	v1.Weekday_WEEKDAY_SATURDAY:  time.Saturday,
}

var protoWeekdayMap = map[time.Weekday]v1.Weekday{
	time.Sunday:    v1.Weekday_WEEKDAY_SUNDAY,
	time.Monday:    v1.Weekday_WEEKDAY_MONDAY,
	time.Tuesday:   v1.Weekday_WEEKDAY_TUESDAY,
	time.Wednesday: v1.Weekday_WEEKDAY_WEDNESDAY,
	time.Thursday:  v1.Weekday_WEEKDAY_THURSDAY,
	time.Friday:    v1.Weekday_WEEKDAY_FRIDAY,
	time.Saturday:  v1.Weekday_WEEKDAY_SATURDAY,
}

var placeSearchFieldMaskToString = map[v1.PlaceSearchFieldMask]string{
	v1.PlaceSearchFieldMask_PLACE_SEARCH_FIELD_MASK_BUSINESS_STATUS:                 "business_status",
	v1.PlaceSearchFieldMask_PLACE_SEARCH_FIELD_MASK_FORMATTED_ADDRESS:               "formatted_address",
	v1.PlaceSearchFieldMask_PLACE_SEARCH_FIELD_MASK_GEOMETRY:                        "geometry",
	v1.PlaceSearchFieldMask_PLACE_SEARCH_FIELD_MASK_GEOMETRY_LOCATION:               "geometry/location",
	v1.PlaceSearchFieldMask_PLACE_SEARCH_FIELD_MASK_GEOMETRY_LOCATION_LAT:           "geometry/location/lat",
	v1.PlaceSearchFieldMask_PLACE_SEARCH_FIELD_MASK_GEOMETRY_LOCATION_LNG:           "geometry/location/lng",
	v1.PlaceSearchFieldMask_PLACE_SEARCH_FIELD_MASK_GEOMETRY_VIEWPORT:               "geometry/viewport",
	v1.PlaceSearchFieldMask_PLACE_SEARCH_FIELD_MASK_GEOMETRY_VIEWPORT_NORTHEAST:     "geometry/viewport/northeast",
	v1.PlaceSearchFieldMask_PLACE_SEARCH_FIELD_MASK_GEOMETRY_VIEWPORT_NORTHEAST_LAT: "geometry/viewport/northeast/lat",
	v1.PlaceSearchFieldMask_PLACE_SEARCH_FIELD_MASK_GEOMETRY_VIEWPORT_NORTHEAST_LNG: "geometry/viewport/northeast/lng",
	v1.PlaceSearchFieldMask_PLACE_SEARCH_FIELD_MASK_GEOMETRY_VIEWPORT_SOUTHWEST:     "geometry/viewport/southwest",
	v1.PlaceSearchFieldMask_PLACE_SEARCH_FIELD_MASK_GEOMETRY_VIEWPORT_SOUTHWEST_LAT: "geometry/viewport/southwest/lat",
	v1.PlaceSearchFieldMask_PLACE_SEARCH_FIELD_MASK_GEOMETRY_VIEWPORT_SOUTHWEST_LNG: "geometry/viewport/southwest/lng",
	v1.PlaceSearchFieldMask_PLACE_SEARCH_FIELD_MASK_ICON:                            "icon",
	v1.PlaceSearchFieldMask_PLACE_SEARCH_FIELD_MASK_ID:                              "id",
	v1.PlaceSearchFieldMask_PLACE_SEARCH_FIELD_MASK_NAME:                            "name",
	v1.PlaceSearchFieldMask_PLACE_SEARCH_FIELD_MASK_OPENING_HOURS:                   "opening_hours",
	v1.PlaceSearchFieldMask_PLACE_SEARCH_FIELD_MASK_OPENING_HOURS_OPEN_NOW:          "opening_hours/open_now",
	v1.PlaceSearchFieldMask_PLACE_SEARCH_FIELD_MASK_PERMANENTLY_CLOSED:              "permanently_closed",
	v1.PlaceSearchFieldMask_PLACE_SEARCH_FIELD_MASK_PHOTOS:                          "photos",
	v1.PlaceSearchFieldMask_PLACE_SEARCH_FIELD_MASK_PLACE_ID:                        "place_id",
	v1.PlaceSearchFieldMask_PLACE_SEARCH_FIELD_MASK_PRICE_LEVEL:                     "price_level",
	v1.PlaceSearchFieldMask_PLACE_SEARCH_FIELD_MASK_RATING:                          "rating",
	v1.PlaceSearchFieldMask_PLACE_SEARCH_FIELD_MASK_USER_RATINGS_TOTAL:              "user_ratings_total",
	v1.PlaceSearchFieldMask_PLACE_SEARCH_FIELD_MASK_VICINITY:                        "vicinity",
	v1.PlaceSearchFieldMask_PLACE_SEARCH_FIELD_MASK_REFERENCE:                       "reference",
	v1.PlaceSearchFieldMask_PLACE_SEARCH_FIELD_MASK_TYPES:                           "types",
}

var inputTypeToString = map[v1.InputType]string{
	v1.InputType_INPUT_TYPE_UNSPECIFIED: "",
	v1.InputType_INPUT_TYPE_TEXT_QUERY:  "textquery",
	v1.InputType_INPUT_TYPE_PHONE_NUMBER: "phonenumber",
}
