package mappers

import (
	v1 "api/src/generated/google_maps/v1"
	"time"

	"googlemaps.github.io/maps"
)

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
		InputType: maps.FindPlaceFromTextInputType(req.InputType.String()),
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
		Geometry:          mapAddressGeometry(c.Geometry),
		Name:              c.Name,
		Icon:              c.Icon,
		PlaceID:           c.PlaceId,
		Rating:            c.Rating,
		UserRatingsTotal:  int(c.UsersRatingTotal),
		Types:             c.Types,
		OpeningHours:      mapOpeningHours(c.OpeningHours),
		Photos:            mapPhotos(c.Photos),
		Vicinity:          c.Vicinity,
		PermanentlyClosed: c.PermanentlyClosed,
		BusinessStatus:    c.BusinessStatus,
		PriceLevel:        int(c.PriceLevel),
	}
}

func mapPhotos(ps []*v1.Photo) []maps.Photo {
	if ps == nil {
		return nil
	}
	photos := make([]maps.Photo, 0, len(ps))
	for _, p := range ps {
		if p != nil {
			photos = append(photos, mapPhoto(p))
		}
	}
	return photos
}

func mapPhoto(p *v1.Photo) maps.Photo {
	return maps.Photo{
		PhotoReference:   p.PhotoReference,
		Height:           int(p.Height),
		Width:            int(p.Width),
		HTMLAttributions: p.HtmlAttributions,
	}
}

func mapOpeningHours(o *v1.OpeningHours) *maps.OpeningHours {
	if o == nil {
		return nil
	}
	return &maps.OpeningHours{
		OpenNow:           &o.OpenNow,
		Periods:           mapOpeningHoursPeriods(o.Periods),
		WeekdayText:       o.WeekdayText,
		PermanentlyClosed: &o.PermanentlyClosed,
	}
}

func mapOpeningHoursPeriods(ps []*v1.OpeningHoursPeriod) []maps.OpeningHoursPeriod {
	if ps == nil {
		return nil
	}
	periods := make([]maps.OpeningHoursPeriod, 0, len(ps))
	for _, p := range ps {
		if p != nil {
			periods = append(periods, maps.OpeningHoursPeriod{
				Open:  mapOpeningHoursOpenClose(p.Open),
				Close: mapOpeningHoursOpenClose(p.Close),
			})
		}
	}
	return periods
}

var weekdayMap = map[v1.Weekday]time.Weekday{
	v1.Weekday_WEEKDAY_SUNDAY:    time.Sunday,
	v1.Weekday_WEEKDAY_MONDAY:    time.Monday,
	v1.Weekday_WEEKDAY_TUESDAY:   time.Tuesday,
	v1.Weekday_WEEKDAY_WEDNESDAY: time.Wednesday,
	v1.Weekday_WEEKDAY_THURSDAY:  time.Thursday,
	v1.Weekday_WEEKDAY_FRIDAY:    time.Friday,
	v1.Weekday_WEEKDAY_SATURDAY:  time.Saturday,
}

func mapOpeningHoursOpenClose(t *v1.OpeningHoursOpenClose) maps.OpeningHoursOpenClose {
	if t == nil {
		return maps.OpeningHoursOpenClose{}
	}
	return maps.OpeningHoursOpenClose{
		Day:  weekdayMap[t.Day],
		Time: t.Time,
	}
}

func mapAddressGeometry(g *v1.AddressGeometry) maps.AddressGeometry {
	if g == nil {
		return maps.AddressGeometry{}
	}
	return maps.AddressGeometry{
		Location:     mapLatLng(g.Location),
		LocationType: g.LocationType,
		Bounds:       mapLatLngBounds(g.Bounds),
		Viewport:     mapLatLngBounds(g.Viewport),
		Types:        g.Types,
	}
}

func mapLatLng(ll *v1.LatLng) maps.LatLng {
	if ll == nil {
		return maps.LatLng{}
	}
	return maps.LatLng{
		Lat: float64(ll.Lat),
		Lng: float64(ll.Lng),
	}
}

func mapLatLngBounds(llb *v1.LatLngBounds) maps.LatLngBounds {
	if llb == nil {
		return maps.LatLngBounds{}
	}
	return maps.LatLngBounds{
		NorthEast: mapLatLng(llb.NorthEast),
		SouthWest: mapLatLng(llb.SouthWest),
	}
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
