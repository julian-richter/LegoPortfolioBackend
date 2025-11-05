package service

import (
	"fmt"
	"net/http"
	"strconv"

	"LegoManagerAPI/internal/config/bricklink"
)

type BricklinkService struct {
	credentials bricklink.BricklinkConfig
	baseURL     string
	httpClient  *http.Client
}

// Common response wrapper
type BricklinkResponse[T any] struct {
	Meta BricklinkMeta `json:"meta"`
	Data T             `json:"data"`
}

type BricklinkMeta struct {
	Description string `json:"description"`
	Message     string `json:"message"`
	Code        int    `json:"code"`
}

// MinifigInfo response
type MinifigInfo struct {
	No           string `json:"no"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	CategoryID   int    `json:"category_id"`
	ImageURL     string `json:"image_url"`
	ThumbnailURL string `json:"thumbnail_url"`
	Weight       string `json:"weight"`
	DimX         string `json:"dim_x"`
	DimY         string `json:"dim_y"`
	DimZ         string `json:"dim_z"`
	YearReleased int    `json:"year_released"`
	IsObsolete   bool   `json:"is_obsolete"`
}

// Subsets Response
type MinifigSubsets []SubsetGroup

type SubsetGroup struct {
	MatchNo int           `json:"match_no"`
	Entries []SubsetEntry `json:"entries"`
}

type SubsetEntry struct {
	Item          SubsetItem `json:"item"`
	ColorID       int        `json:"color_id"`
	Quantity      int        `json:"quantity"`
	ExtraQuantity int        `json:"extra_quantity"`
	IsAlternate   bool       `json:"is_alternate"`
	IsCounterpart bool       `json:"is_counterpart"`
}

type SubsetItem struct {
	No         string `json:"no"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	CategoryID int    `json:"category_id"`
}

// Price response
type MinifigPrice struct {
	Item          PriceItem     `json:"item"`
	NewOrUsed     string        `json:"new_or_used"`
	CurrencyCode  string        `json:"currency_code"`
	MinPrice      string        `json:"min_price"`
	MaxPrice      string        `json:"max_price"`
	AvgPrice      string        `json:"avg_price"`
	QtyAvgPrice   string        `json:"qty_avg_price"`
	UnitQuantity  int           `json:"unit_quantity"`
	TotalQuantity int           `json:"total_quantity"`
	PriceDetail   []PriceDetail `json:"price_detail"`
}

type PriceItem struct {
	No   string `json:"no"`
	Type string `json:"type"`
}

type PriceDetail struct {
	Quantity          int    `json:"quantity"`
	UnitPrice         string `json:"unit_price"`
	ShippingAvailable bool   `json:"shipping_available"`
}

// Better structured combined response
type MinifigCompleteResponse struct {
	MinifigID  string            `json:"minifig_id"`
	BasicInfo  MinifigBasicInfo  `json:"basic_info"`
	Components MinifigComponents `json:"components"`
	Market     MinifigMarketData `json:"market_data"`
	Images     MinifigImages     `json:"images"`
	Metadata   ResponseMetadata  `json:"metadata"`
}

type MinifigBasicInfo struct {
	Name         string     `json:"name"`
	Type         string     `json:"type"`
	CategoryID   int        `json:"category_id"`
	YearReleased int        `json:"year_released"`
	IsObsolete   bool       `json:"is_obsolete"`
	Dimensions   Dimensions `json:"dimensions"`
}

type Dimensions struct {
	Weight string `json:"weight_grams"`
	Length string `json:"length_cm"`
	Width  string `json:"width_cm"`
	Height string `json:"height_cm"`
}

type MinifigComponents struct {
	TotalParts int             `json:"total_parts"`
	Parts      []ComponentPart `json:"parts"`
}

type ComponentPart struct {
	PartNumber  string `json:"part_number"`
	PartName    string `json:"part_name"`
	PartType    string `json:"part_type"`
	ColorID     int    `json:"color_id"`
	Quantity    int    `json:"quantity"`
	IsAlternate bool   `json:"is_alternate"`
	CategoryID  int    `json:"category_id"`
}

type MinifigMarketData struct {
	Currency       string                `json:"currency"`
	Condition      string                `json:"condition"`
	PriceSummary   PriceSummary          `json:"price_summary"`
	Availability   AvailabilitySummary   `json:"availability"`
	PriceBreakdown []PriceBreakdownEntry `json:"price_breakdown"`
}

type PriceSummary struct {
	Minimum         float64 `json:"minimum_usd"`
	Maximum         float64 `json:"maximum_usd"`
	Average         float64 `json:"average_usd"`
	WeightedAverage float64 `json:"weighted_average_usd"`
}

type AvailabilitySummary struct {
	TotalListings   int `json:"total_listings"`
	TotalQuantity   int `json:"total_quantity_available"`
	WithShipping    int `json:"listings_with_shipping"`
	WithoutShipping int `json:"listings_without_shipping"`
}

type PriceBreakdownEntry struct {
	Quantity          int     `json:"quantity"`
	PricePerUnit      float64 `json:"price_per_unit_usd"`
	ShippingAvailable bool    `json:"shipping_available"`
}

type MinifigImages struct {
	FullSize  string `json:"full_size_url"`
	Thumbnail string `json:"thumbnail_url"`
}

type ResponseMetadata struct {
	FetchedAt        string          `json:"fetched_at"`
	TotalFetchTimeMs int64           `json:"total_fetch_time_ms"`
	EndpointTimings  EndpointTimings `json:"endpoint_timings_ms"`
	DataSources      []string        `json:"data_sources"`
}

type EndpointTimings struct {
	BasicInfo  int64 `json:"basic_info"`
	Components int64 `json:"components"`
	MarketData int64 `json:"market_data"`
}

type MinifigComplete struct {
	Info                  *MinifigInfo     `json:"info"`
	Subsets               MinifigSubsets   `json:"subsets"`
	Price                 *MinifigPrice    `json:"price"`
	FetchTimeMs           int64            `json:"fetch_time_ms"`
	IndividualFetchTimeMs map[string]int64 `json:"individual_fetch_time_ms"`
}

// Helper to convert raw response to structured response
func (mc *MinifigComplete) ToStructuredResponse() *MinifigCompleteResponse {
	// Extract basic info
	basicInfo := MinifigBasicInfo{
		Name:         mc.Info.Name,
		Type:         mc.Info.Type,
		CategoryID:   mc.Info.CategoryID,
		YearReleased: mc.Info.YearReleased,
		IsObsolete:   mc.Info.IsObsolete,
		Dimensions: Dimensions{
			Weight: mc.Info.Weight,
			Length: mc.Info.DimX,
			Width:  mc.Info.DimY,
			Height: mc.Info.DimZ,
		},
	}

	// Extract components
	var parts []ComponentPart
	totalParts := 0
	for _, group := range mc.Subsets {
		for _, entry := range group.Entries {
			parts = append(parts, ComponentPart{
				PartNumber:  entry.Item.No,
				PartName:    entry.Item.Name,
				PartType:    entry.Item.Type,
				ColorID:     entry.ColorID,
				Quantity:    entry.Quantity,
				IsAlternate: entry.IsAlternate,
				CategoryID:  entry.Item.CategoryID,
			})
			totalParts += entry.Quantity
		}
	}

	components := MinifigComponents{
		TotalParts: totalParts,
		Parts:      parts,
	}

	// Extract market data with proper float parsing
	minPrice, _ := strconv.ParseFloat(mc.Price.MinPrice, 64)
	maxPrice, _ := strconv.ParseFloat(mc.Price.MaxPrice, 64)
	avgPrice, _ := strconv.ParseFloat(mc.Price.AvgPrice, 64)
	qtyAvgPrice, _ := strconv.ParseFloat(mc.Price.QtyAvgPrice, 64)

	var priceBreakdown []PriceBreakdownEntry
	withShipping := 0
	withoutShipping := 0

	for _, detail := range mc.Price.PriceDetail {
		price, _ := strconv.ParseFloat(detail.UnitPrice, 64)
		priceBreakdown = append(priceBreakdown, PriceBreakdownEntry{
			Quantity:          detail.Quantity,
			PricePerUnit:      price,
			ShippingAvailable: detail.ShippingAvailable,
		})

		if detail.ShippingAvailable {
			withShipping++
		} else {
			withoutShipping++
		}
	}

	marketData := MinifigMarketData{
		Currency:  mc.Price.CurrencyCode,
		Condition: mc.Price.NewOrUsed,
		PriceSummary: PriceSummary{
			Minimum:         minPrice,
			Maximum:         maxPrice,
			Average:         avgPrice,
			WeightedAverage: qtyAvgPrice,
		},
		Availability: AvailabilitySummary{
			TotalListings:   mc.Price.UnitQuantity,
			TotalQuantity:   mc.Price.TotalQuantity,
			WithShipping:    withShipping,
			WithoutShipping: withoutShipping,
		},
		PriceBreakdown: priceBreakdown,
	}

	// Fix image URLs (add https:)
	imageURL := mc.Info.ImageURL
	thumbnailURL := mc.Info.ThumbnailURL
	if imageURL != "" && imageURL[:2] == "//" {
		imageURL = "https:" + imageURL
	}
	if thumbnailURL != "" && thumbnailURL[:2] == "//" {
		thumbnailURL = "https:" + thumbnailURL
	}

	images := MinifigImages{
		FullSize:  imageURL,
		Thumbnail: thumbnailURL,
	}

	// Metadata
	metadata := ResponseMetadata{
		FetchedAt:        fmt.Sprintf("%d", mc.FetchTimeMs),
		TotalFetchTimeMs: mc.FetchTimeMs,
		EndpointTimings: EndpointTimings{
			BasicInfo:  mc.IndividualFetchTimeMs["info"],
			Components: mc.IndividualFetchTimeMs["subsets"],
			MarketData: mc.IndividualFetchTimeMs["price"],
		},
		DataSources: []string{"Bricklink API v1"},
	}

	return &MinifigCompleteResponse{
		MinifigID:  mc.Info.No,
		BasicInfo:  basicInfo,
		Components: components,
		Market:     marketData,
		Images:     images,
		Metadata:   metadata,
	}
}
