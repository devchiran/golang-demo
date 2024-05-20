package catelog

import "errors"

var ErrNotFound = errors.New("not found")
var ErrPublisherMismatch = errors.New("publisher_slug mismatch")
var ErrInvalidDate = errors.New("invalid date. Format should be yyyy-mm-dd")
var ErrMissingLabel = errors.New("label must be provided in request body")
var ErrMissingLocation = errors.New("location must be provided in request body")
var ErrInvalidCreatorLimit = errors.New("creator_limit must be greater than 0 and less than 1000")
var ErrInvalidChallengeType = errors.New("invalid challenge type")
var ErrAssignNonOpenChallenge = errors.New("can only assign creators to open challenges")

var ErrInvalidStatus = errors.New("status must be draft or submitted")
