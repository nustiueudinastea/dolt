#!/usr/bin/env bats
load $BATS_TEST_DIRNAME/helper/common.bash

setup() {
    setup_common

    dolt sql -q "create table tbl (i int auto_increment primary key, guid char(36))"
    dolt commit -A -m "create tbl"

    make_inserts
}

teardown() {
    assert_feature_version
    teardown_common
}

# Insert 25 new rows, then commit.
make_inserts() {
  for ((i=1; i<=25; i++))
  do
    dolt sql -q "INSERT INTO tbl (guid) VALUES (UUID())"
  done
  dolt commit -a -m "Add 25 values"
}

# Randomly update 10 rows, then commit.
make_updates() {
  for ((i=1; i<=10; i++))
  do
        dolt sql -q "
    SET @max_id = (SELECT MAX(i) FROM tbl);
    SET @random_id = FLOOR(1 + RAND() * @max_id);
    UPDATE tbl SET guid = UUID() WHERE i >= @random_id LIMIT 1;"
  done
  dolt commit -a -m "Update 10 values."
}

@test "archive: too few chunks" {
  make_updates
  dolt gc

  run dolt admin archive
  [ "$status" -eq 1 ]
  [[ "$output" =~ "Not enough samples to build default dictionary" ]] || false
}

# This test runs over 45 seconds, resulting in a timeout in lambdabats
# bats test_tags=no_lambda
@test "archive: single archive" {
  # We need at least 25 chunks to create an archive.
  for ((j=1; j<=10; j++))
  do
    make_updates
    make_inserts
  done

  dolt gc
  dolt admin archive

  files=$(find . -name "*darc" | wc -l | sed 's/[ \t]//g')
  [ "$files" -eq "1" ]

  # Ensure updates continue to work.
  make_updates
}


# This test runs over 45 seconds, resulting in a timeout in lambdabats
# bats test_tags=no_lambda
@test "archive: multiple archives" {
  # We need at least 25 chunks to create an archive.
  for ((j=1; j<=10; j++))
  do
    make_updates
    make_inserts
  done
  dolt gc

  for ((j=1; j<=10; j++))
  do
    make_updates
    make_inserts
  done
  dolt gc

  for ((j=1; j<=10; j++))
  do
    make_updates
    make_inserts
  done
  dolt gc

  dolt admin archive

  files=$(find . -name "*darc" | wc -l | sed 's/[ \t]//g')
  [ "$files" -eq "3" ]

  # dolt log --stat will load every single chunk.
  commits=$(dolt log --stat --oneline | wc -l | sed 's/[ \t]//g')
  [ "$commits" -eq "186" ]
}

