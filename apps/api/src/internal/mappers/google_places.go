package mappers

import (
	"log/slog"
	"strings"

	v1 "api/src/generated/google_maps/v1"

	"cloud.google.com/go/maps/places/apiv1/placespb"
)

func boolPtr(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

func int32Ptr(i *int32) int32 {
	if i == nil {
		return 0
	}
	return *i
}

func stringPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

var businessStatusMap = map[placespb.Place_BusinessStatus]v1.BusinessStatus{
	placespb.Place_BUSINESS_STATUS_UNSPECIFIED: v1.BusinessStatus_BUSINESS_STATUS_UNSPECIFIED,
	placespb.Place_OPERATIONAL:                 v1.BusinessStatus_BUSINESS_STATUS_OPERATIONAL,
	placespb.Place_CLOSED_TEMPORARILY:          v1.BusinessStatus_BUSINESS_STATUS_CLOSED_TEMPORARILY,
	placespb.Place_CLOSED_PERMANENTLY:          v1.BusinessStatus_BUSINESS_STATUS_CLOSED_PERMANENTLY,
}

var priceLevelMap = map[placespb.PriceLevel]v1.PriceLevel{
	placespb.PriceLevel_PRICE_LEVEL_UNSPECIFIED:    v1.PriceLevel_PRICE_LEVEL_UNSPECIFIED,
	placespb.PriceLevel_PRICE_LEVEL_FREE:           v1.PriceLevel_PRICE_LEVEL_FREE,
	placespb.PriceLevel_PRICE_LEVEL_INEXPENSIVE:    v1.PriceLevel_PRICE_LEVEL_INEXPENSIVE,
	placespb.PriceLevel_PRICE_LEVEL_MODERATE:       v1.PriceLevel_PRICE_LEVEL_MODERATE,
	placespb.PriceLevel_PRICE_LEVEL_EXPENSIVE:      v1.PriceLevel_PRICE_LEVEL_EXPENSIVE,
	placespb.PriceLevel_PRICE_LEVEL_VERY_EXPENSIVE: v1.PriceLevel_PRICE_LEVEL_VERY_EXPENSIVE,
}

var priceLevelReverseMap = map[v1.PriceLevel]placespb.PriceLevel{
	v1.PriceLevel_PRICE_LEVEL_UNSPECIFIED:    placespb.PriceLevel_PRICE_LEVEL_UNSPECIFIED,
	v1.PriceLevel_PRICE_LEVEL_FREE:           placespb.PriceLevel_PRICE_LEVEL_FREE,
	v1.PriceLevel_PRICE_LEVEL_INEXPENSIVE:    placespb.PriceLevel_PRICE_LEVEL_INEXPENSIVE,
	v1.PriceLevel_PRICE_LEVEL_MODERATE:       placespb.PriceLevel_PRICE_LEVEL_MODERATE,
	v1.PriceLevel_PRICE_LEVEL_EXPENSIVE:      placespb.PriceLevel_PRICE_LEVEL_EXPENSIVE,
	v1.PriceLevel_PRICE_LEVEL_VERY_EXPENSIVE: placespb.PriceLevel_PRICE_LEVEL_VERY_EXPENSIVE,
}

func BuildFieldMask(requestedFields []string) string {
	if len(requestedFields) == 0 {
		return "*"
	}
	return strings.Join(requestedFields, ",")
}

func PriceLevelsToSDK(levels []v1.PriceLevel) []placespb.PriceLevel {
	result := make([]placespb.PriceLevel, len(levels))
	for i, level := range levels {
		if sdkLevel, exists := priceLevelReverseMap[level]; exists {
			result[i] = sdkLevel
		} else {
			result[i] = placespb.PriceLevel(level)
		}
	}
	return result
}

func SearchTextResponseToProto(resp *placespb.SearchTextResponse) *v1.SearchTextResponse {
	if resp == nil {
		return nil
	}

	places := make([]*v1.Place, 0, len(resp.Places))
	for _, place := range resp.Places {
		convertedPlace := PlaceToProto(place)
		places = append(places, convertedPlace)
	}

	return &v1.SearchTextResponse{
		Places: places,
	}
}

func AutocompletePlacesResponseToProto(resp *placespb.AutocompletePlacesResponse) *v1.AutocompletePlacesResponse {
	if resp == nil {
		return nil
	}
	result := &v1.AutocompletePlacesResponse{Suggestions: suggestionsToProto(resp.Suggestions)}
	return result
}

func suggestionsToProto(suggestions []*placespb.AutocompletePlacesResponse_Suggestion) []*v1.Suggestion {
	if len(suggestions) == 0 {
		return nil
	}

	result := make([]*v1.Suggestion, 0, len(suggestions))
	for _, suggestion := range suggestions {
		if suggestion == nil {
			continue
		}

		var placePrediction *v1.PlacePrediction
		var queryPrediction *v1.QueryPrediction

		switch suggestion.Kind.(type) {
		case *placespb.AutocompletePlacesResponse_Suggestion_PlacePrediction_:
			if placePred := suggestion.GetPlacePrediction(); placePred != nil {
				placePrediction = PlacePredictionToProto(placePred)
			}
		case *placespb.AutocompletePlacesResponse_Suggestion_QueryPrediction_:
			if queryPred := suggestion.GetQueryPrediction(); queryPred != nil {
				queryPrediction = QueryPredictionToProto(queryPred)
			}
		}

		result = append(result, &v1.Suggestion{
			PlacePrediction: placePrediction,
			QueryPrediction: queryPrediction,
		})
	}
	return result
}

func PlacePredictionToProto(prediction *placespb.AutocompletePlacesResponse_Suggestion_PlacePrediction) *v1.PlacePrediction {
	if prediction == nil {
		return nil
	}
	return &v1.PlacePrediction{
		Place:   prediction.Place,
		PlaceId: prediction.PlaceId,
		Text:    formattableTextToProto(prediction.Text),
		StructuredFormat: &v1.StructuredFormat{
			MainText:      formattableTextToProto(prediction.StructuredFormat.MainText),
			SecondaryText: formattableTextToProto(prediction.StructuredFormat.SecondaryText),
		},
		Types:          prediction.Types,
		DistanceMeters: prediction.DistanceMeters,
	}
}

func QueryPredictionToProto(prediction *placespb.AutocompletePlacesResponse_Suggestion_QueryPrediction) *v1.QueryPrediction {
	if prediction == nil {
		return nil
	}
	slog.Info("QueryPredictionToProto", "prediction", prediction)
	return &v1.QueryPrediction{
		Text: formattableTextToProto(prediction.Text),
		StructuredFormat: &v1.StructuredFormat{
			MainText:      formattableTextToProto(prediction.StructuredFormat.MainText),
			SecondaryText: formattableTextToProto(prediction.StructuredFormat.SecondaryText),
		},
	}
}

func formattableTextToProto(text *placespb.AutocompletePlacesResponse_Suggestion_FormattableText) *v1.FormattableText {
	if text == nil {
		return nil
	}
	return &v1.FormattableText{
		Text:    text.Text,
		Matches: stringRangesToProto(text.Matches),
	}
}

func stringRangesToProto(ranges []*placespb.AutocompletePlacesResponse_Suggestion_StringRange) []*v1.StringRange {
	if len(ranges) == 0 {
		return nil
	}

	result := make([]*v1.StringRange, 0, len(ranges))
	for _, r := range ranges {
		if r == nil {
			continue
		}
		result = append(result, &v1.StringRange{
			StartOffset: r.StartOffset,
			EndOffset:   r.EndOffset,
		})
	}
	return result
}

func PlaceToProto(place *placespb.Place) *v1.Place {
	if place == nil {
		return nil
	}

	result := &v1.Place{
		Name:                     place.Name,
		Id:                       place.Id,
		Types:                    place.Types,
		PrimaryType:              place.PrimaryType,
		NationalPhoneNumber:      place.NationalPhoneNumber,
		InternationalPhoneNumber: place.InternationalPhoneNumber,
		FormattedAddress:         place.FormattedAddress,
		ShortFormattedAddress:    place.ShortFormattedAddress,
		Rating:                   place.Rating,
		GoogleMapsUri:            place.GoogleMapsUri,
		WebsiteUri:               place.WebsiteUri,
		AdrFormatAddress:         place.AdrFormatAddress,
		BusinessStatus:           mapBusinessStatus(place.BusinessStatus),
		PriceLevel:               mapPriceLevel(place.PriceLevel),
		IconMaskBaseUri:          place.IconMaskBaseUri,
		IconBackgroundColor:      place.IconBackgroundColor,
		UserRatingCount:          int32Ptr(place.UserRatingCount),
		Takeout:                  boolPtr(place.Takeout),
		Delivery:                 boolPtr(place.Delivery),
		DineIn:                   boolPtr(place.DineIn),
		CurbsidePickup:           boolPtr(place.CurbsidePickup),
		Reservable:               boolPtr(place.Reservable),
		ServesBreakfast:          boolPtr(place.ServesBreakfast),
		ServesLunch:              boolPtr(place.ServesLunch),
		ServesDinner:             boolPtr(place.ServesDinner),
		ServesBeer:               boolPtr(place.ServesBeer),
		ServesWine:               boolPtr(place.ServesWine),
		ServesBrunch:             boolPtr(place.ServesBrunch),
		ServesVegetarianFood:     boolPtr(place.ServesVegetarianFood),
		OutdoorSeating:           boolPtr(place.OutdoorSeating),
		LiveMusic:                boolPtr(place.LiveMusic),
		MenuForChildren:          boolPtr(place.MenuForChildren),
		ServesCocktails:          boolPtr(place.ServesCocktails),
		ServesDessert:            boolPtr(place.ServesDessert),
		ServesCoffee:             boolPtr(place.ServesCoffee),
		GoodForChildren:          boolPtr(place.GoodForChildren),
		AllowsDogs:               boolPtr(place.AllowsDogs),
		Restroom:                 boolPtr(place.Restroom),
		GoodForGroups:            boolPtr(place.GoodForGroups),
		GoodForWatchingSports:    boolPtr(place.GoodForWatchingSports),
		PureServiceAreaBusiness:  boolPtr(place.PureServiceAreaBusiness),
		UtcOffsetMinutes:         int32Ptr(place.UtcOffsetMinutes),
	}

	if place.DisplayName != nil {
		result.DisplayName = &v1.LocalizedText{
			Text:         place.DisplayName.Text,
			LanguageCode: place.DisplayName.LanguageCode,
		}
	}

	if place.PrimaryTypeDisplayName != nil {
		result.PrimaryTypeDisplayName = &v1.LocalizedText{
			Text:         place.PrimaryTypeDisplayName.Text,
			LanguageCode: place.PrimaryTypeDisplayName.LanguageCode,
		}
	}

	result.Photos = photosToProto(place.Photos)
	result.Attributions = attributionsToProto(place.Attributions)

	return result
}

func mapBusinessStatus(status placespb.Place_BusinessStatus) v1.BusinessStatus {
	if mapped, exists := businessStatusMap[status]; exists {
		return mapped
	}
	return v1.BusinessStatus_BUSINESS_STATUS_UNSPECIFIED
}

func mapPriceLevel(level placespb.PriceLevel) v1.PriceLevel {
	if mapped, exists := priceLevelMap[level]; exists {
		return mapped
	}
	return v1.PriceLevel_PRICE_LEVEL_UNSPECIFIED
}

func photosToProto(photos []*placespb.Photo) []*v1.Photo {
	if len(photos) == 0 {
		return nil
	}

	result := make([]*v1.Photo, 0, len(photos))
	for _, photo := range photos {
		if photo == nil {
			continue
		}

		authorAttributions := make([]string, 0, len(photo.AuthorAttributions))
		for _, attr := range photo.AuthorAttributions {
			if attr != nil {
				authorAttributions = append(authorAttributions, attr.Uri)
			}
		}

		result = append(result, &v1.Photo{
			Name:               photo.Name,
			WidthPx:            photo.WidthPx,
			HeightPx:           photo.HeightPx,
			AuthorAttributions: authorAttributions,
		})
	}
	return result
}

func attributionsToProto(attributions []*placespb.Place_Attribution) []*v1.Attribution {
	if len(attributions) == 0 {
		return nil
	}

	result := make([]*v1.Attribution, 0, len(attributions))
	for _, attr := range attributions {
		if attr == nil {
			continue
		}
		result = append(result, &v1.Attribution{
			Provider:    attr.Provider,
			ProviderUri: attr.ProviderUri,
		})
	}
	return result
}
