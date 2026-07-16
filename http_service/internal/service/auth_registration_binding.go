/*
 * Registration residence binding workflow.
 * 1. Validate the optional building, floor, and unit selection as one request.
 * 2. Resolve the POS unit ID required by the iSmart OwnerReg API.
 * 3. Persist only a pending application after iSmart accepts the request.
 */
package service

import (
	"context"
	"strings"

	"gorm.io/gorm"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/model"
)

const residenceBindingStatusPending = "pending"

// 1. validateRegistrationResidenceBinding validates optional residence selection completeness.
func validateRegistrationResidenceBinding(buildingID string, floor string, unit string) (bool, error) {
	requested := strings.TrimSpace(buildingID) != "" || strings.TrimSpace(floor) != "" || strings.TrimSpace(unit) != ""
	if !requested {
		return false, nil
	}
	if strings.TrimSpace(buildingID) == "" || strings.TrimSpace(floor) == "" || strings.TrimSpace(unit) == "" {
		return false, errcode.New(errcode.CodeValidationError, "building, floor, and unit are required together")
	}

	return true, nil
}

// 2. submitRegistrationResidenceBinding sends the OwnerReg application after the local account is committed.
func (s *AuthService) submitRegistrationResidenceBinding(
	ctx context.Context,
	userID int64,
	buildingID string,
	buildingName string,
	floor string,
	unit string,
	phone string,
	email string,
	engName string,
	params EmailPasswordParams,
) error {
	unitID, err := s.resolveRegistrationResidenceUnit(ctx, buildingID, floor, unit)
	if err != nil {
		return registrationResidenceBindingError(err)
	}

	isReceiveEmail := true
	if params.IsReceiveEmail != nil {
		isReceiveEmail = *params.IsReceiveEmail
	}
	err = NewIsmartExternalService(s.runtime).SubmitOwnerBindingRequestAccepted(ctx, userID, IsmartOwnerBindingParams{
		BuildingID:        buildingID,
		OwnedFlat:         []string{unitID},
		Role:              "業主",
		OwnerNote:         strings.TrimSpace(params.Remark),
		IsReceiveEmail:    &isReceiveEmail,
		RegistrationTel:   phone,
		RegistrationEmail: email,
		ClientName:        engName,
		ClientIDCard:      strings.TrimSpace(params.IDCard),
		ClientTel:         phone,
	})
	if err != nil {
		return registrationResidenceBindingError(err)
	}

	return s.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		community, err := s.resolveRegistrationCommunity(ctx, tx, strings.TrimSpace(buildingID), strings.TrimSpace(buildingName))
		if err != nil {
			return err
		}
		updates := map[string]any{
			"primary_community_id":     community.ID,
			"residence_floor":          strings.TrimSpace(floor),
			"residence_unit":           strings.TrimSpace(unit),
			"residence_binding_status": residenceBindingStatusPending,
			"district_code":            community.DistrictCode,
		}
		if err := tx.WithContext(ctx).Model(&model.UserProfile{}).Where("user_id = ?", userID).Updates(updates).Error; err != nil {
			return errcode.New(errcode.CodeInternalError, "failed to save property binding request")
		}

		return nil
	})
}

// 3. resolveRegistrationResidenceUnit resolves the displayed registration unit to one iSmart unit ID.
func (s *AuthService) resolveRegistrationResidenceUnit(ctx context.Context, buildingID string, floor string, unit string) (string, error) {
	units, err := NewPOSBuildingService(s.runtime).ListUnits(ctx, strings.TrimSpace(buildingID))
	if err != nil {
		return "", err
	}

	targetFloor := strings.TrimSpace(floor)
	targetUnit := strings.TrimSpace(unit)
	matches := make([]string, 0, 1)
	for _, item := range units {
		unitID := strings.TrimSpace(firstPOSUnitValue(item.UnitID, item.ID))
		if unitID == "" {
			continue
		}
		if unitID == targetUnit || (strings.TrimSpace(item.Floor) == targetFloor && strings.TrimSpace(firstPOSUnitValue(item.Unit, item.UnitName, item.Name)) == targetUnit) {
			matches = append(matches, unitID)
		}
	}
	if len(matches) != 1 {
		return "", errcode.New(errcode.CodeValidationError, "selected building unit is invalid")
	}

	return matches[0], nil
}

// 4. firstPOSUnitValue returns the first populated POS unit representation.
func firstPOSUnitValue(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}

	return ""
}

// 5. registrationResidenceBindingError keeps the recoverable account-created state explicit to callers.
func registrationResidenceBindingError(error) error {
	return errcode.New(errcode.CodeInternalError, "account created but property binding request failed")
}
