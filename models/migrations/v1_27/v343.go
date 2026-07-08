// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package v1_27

import (
	"gitea.dev/models/db"

	"xorm.io/xorm"
)

type repositoryReticulumThanks struct {
	NumReticulumThanks int `xorm:"NOT NULL DEFAULT 0"`
}

func (repositoryReticulumThanks) TableName() string {
	return "repository"
}

func AddRepositoryReticulumThanks(x db.EngineMigration) error {
	_, err := x.SyncWithOptions(xorm.SyncOptions{
		IgnoreDropIndices: true,
		IgnoreConstrains:  true,
	}, new(repositoryReticulumThanks))
	return err
}
