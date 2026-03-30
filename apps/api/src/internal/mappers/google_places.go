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

// regionToName maps ISO 3166-1 alpha-2 codes to English country names for consistent storage.
var regionToName = map[string]string{
	"AF": "Afghanistan", "AL": "Albania", "DZ": "Algeria", "AD": "Andorra",
	"AO": "Angola", "AR": "Argentina", "AM": "Armenia", "AU": "Australia",
	"AT": "Austria", "AZ": "Azerbaijan", "BH": "Bahrain", "BD": "Bangladesh",
	"BY": "Belarus", "BE": "Belgium", "BZ": "Belize", "BJ": "Benin",
	"BT": "Bhutan", "BO": "Bolivia", "BA": "Bosnia and Herzegovina",
	"BW": "Botswana", "BR": "Brazil", "BN": "Brunei", "BG": "Bulgaria",
	"BF": "Burkina Faso", "BI": "Burundi", "KH": "Cambodia", "CM": "Cameroon",
	"CA": "Canada", "CF": "Central African Republic", "TD": "Chad",
	"CL": "Chile", "CN": "China", "CO": "Colombia", "CG": "Congo",
	"CD": "DR Congo", "CR": "Costa Rica", "HR": "Croatia", "CU": "Cuba",
	"CY": "Cyprus", "CZ": "Czech Republic", "DK": "Denmark", "DJ": "Djibouti",
	"DO": "Dominican Republic", "EC": "Ecuador", "EG": "Egypt",
	"SV": "El Salvador", "EE": "Estonia", "ET": "Ethiopia", "FI": "Finland",
	"FR": "France", "GA": "Gabon", "GE": "Georgia", "DE": "Germany",
	"GH": "Ghana", "GR": "Greece", "GT": "Guatemala", "GN": "Guinea",
	"HT": "Haiti", "HN": "Honduras", "HU": "Hungary", "IS": "Iceland",
	"IN": "India", "ID": "Indonesia", "IR": "Iran", "IQ": "Iraq",
	"IE": "Ireland", "IL": "Israel", "IT": "Italy", "JM": "Jamaica",
	"JP": "Japan", "JO": "Jordan", "KZ": "Kazakhstan", "KE": "Kenya",
	"KW": "Kuwait", "KG": "Kyrgyzstan", "LA": "Laos", "LV": "Latvia",
	"LB": "Lebanon", "LY": "Libya", "LI": "Liechtenstein", "LT": "Lithuania",
	"LU": "Luxembourg", "MK": "North Macedonia", "MG": "Madagascar",
	"MY": "Malaysia", "MV": "Maldives", "ML": "Mali", "MT": "Malta",
	"MX": "Mexico", "MD": "Moldova", "MC": "Monaco", "MN": "Mongolia",
	"ME": "Montenegro", "MA": "Morocco", "MZ": "Mozambique", "MM": "Myanmar",
	"NA": "Namibia", "NP": "Nepal", "NL": "Netherlands", "NZ": "New Zealand",
	"NI": "Nicaragua", "NE": "Niger", "NG": "Nigeria", "NO": "Norway",
	"OM": "Oman", "PK": "Pakistan", "PA": "Panama", "PY": "Paraguay",
	"PE": "Peru", "PH": "Philippines", "PL": "Poland", "PT": "Portugal",
	"QA": "Qatar", "RO": "Romania", "RU": "Russia", "RW": "Rwanda",
	"SA": "Saudi Arabia", "SN": "Senegal", "RS": "Serbia", "SL": "Sierra Leone",
	"SG": "Singapore", "SK": "Slovakia", "SI": "Slovenia", "SO": "Somalia",
	"ZA": "South Africa", "SS": "South Sudan", "ES": "Spain", "LK": "Sri Lanka",
	"SD": "Sudan", "SE": "Sweden", "CH": "Switzerland", "SY": "Syria",
	"TW": "Taiwan", "TJ": "Tajikistan", "TZ": "Tanzania", "TH": "Thailand",
	"TN": "Tunisia", "TR": "Turkey", "TM": "Turkmenistan", "UG": "Uganda",
	"UA": "Ukraine", "AE": "United Arab Emirates", "GB": "United Kingdom",
	"US": "United States", "UY": "Uruguay", "UZ": "Uzbekistan",
	"VE": "Venezuela", "VN": "Vietnam", "YE": "Yemen", "ZM": "Zambia",
	"ZW": "Zimbabwe",
}

func regionCodeToCountry(code string) string {
	if name, ok := regionToName[code]; ok {
		return name
	}
	return code
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

	if place.PostalAddress != nil {
		country := regionCodeToCountry(place.PostalAddress.RegionCode)
		result.PostalAddress = &v1.PostalAddress{
			Locality:           place.PostalAddress.Locality,
			AdministrativeArea: place.PostalAddress.AdministrativeArea,
			Country:            country,
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
