"""Raw storage test (umbrella doc §3.6 / ADR-0025 봉투 암호화)."""

from __future__ import annotations

import os

import pytest

from backend_knowledge.storage import RawStore, get_raw_store
from backend_knowledge.storage.raw_store import (
    InvalidSourceNameError,
    SOURCE_NAME_PATTERN,
)


class TestRawStorePlaintext:
    """Test raw store in plaintext mode (no KEK set, PoC default)."""

    def test_save_and_load(
        self, temp_var_dir
    ) -> None:
        """Save raw then load should return same body + sha256."""
        store = get_raw_store()
        result = store.save(
            source="test_source",
            type_="dataset",
            name="test_dataset_001",
            body="# Test\n\nBody content",
            registered_by="u_test_001",
        )
        assert result.envelope_encrypted is False  # plaintext mode
        assert result.sha256
        assert result.size > 0

        record = store.load(source="test_source", raw_id=result.raw_id)
        assert record.body == "# Test\n\nBody content"
        assert record.sha256 == result.sha256
        assert record.registered_by == "u_test_001"  # from sidecar (Codex P1 fix)

    def test_list_source(self, temp_var_dir) -> None:
        """list_source should return saved raw_ids."""
        store = get_raw_store()
        store.save(source="test_source", type_="dataset", name="a", body="a")
        store.save(source="test_source", type_="dataset", name="b", body="b")
        store.save(source="test_source", type_="dataset", name="c", body="c")

        ids = store.list_source(source="test_source")
        assert len(ids) == 3

    def test_exists(self, temp_var_dir) -> None:
        """exists should return True for saved raws, False for non-existent."""
        store = get_raw_store()
        result = store.save(source="test_source", type_="dataset", name="x", body="x")
        assert store.exists(source="test_source", raw_id=result.raw_id) is True
        assert store.exists(source="test_source", raw_id="nonexistent-id") is False
        assert store.exists(source="other_source", raw_id=result.raw_id) is False


class TestRawStoreEncrypted:
    """Test raw store in encrypted mode (KEK set)."""

    def test_save_and_load_encrypted(self, temp_var_dir) -> None:
        """Save + load with KEK should decrypt correctly."""
        import os
        kek = os.urandom(32)
        store = RawStore(base_dir=temp_var_dir / "raw", kek=kek)
        result = store.save(
            source="encrypted_source",
            type_="metric",
            name="encrypted_metric",
            body="secret body",
            registered_by="u_test",
        )
        assert result.envelope_encrypted is True

        record = store.load(source="encrypted_source", raw_id=result.raw_id)
        assert record.body == "secret body"

    def test_wrong_kek_fails_decryption(self, temp_var_dir) -> None:
        """Decryption with wrong KEK should fail."""
        import os
        kek_correct = os.urandom(32)
        kek_wrong = os.urandom(32)
        store_correct = RawStore(base_dir=temp_var_dir / "raw", kek=kek_correct)
        result = store_correct.save(
            source="encrypted_source",
            type_="dataset",
            name="x",
            body="secret",
        )
        # Try to load with wrong KEK
        store_wrong = RawStore(base_dir=temp_var_dir / "raw", kek=kek_wrong)
        with pytest.raises(Exception):  # cryptography raises InvalidTag
            store_wrong.load(source="encrypted_source", raw_id=result.raw_id)

    def test_kek_length_must_be_32(self, temp_var_dir) -> None:
        """KEK with wrong length should raise RawStoreError."""
        with pytest.raises(Exception, match="KEK must be 32 bytes"):
            RawStore(base_dir=temp_var_dir / "raw", kek=os.urandom(16))


class TestRawMetadataSidecar:
    """Codex P1 review fix: persist raw metadata with each raw body."""

    def test_metadata_sidecar_persisted(
        self, temp_var_dir
    ) -> None:
        """After save(), .meta.json should contain type/name/owner/visibility."""
        import json
        store = get_raw_store()
        result = store.save(
            source="meta_test",
            type_="dataset",
            name="users_2026",
            body="# Users",
            registered_by="u_owner",
            owner_org_id="ou_test_dept_a",
            owner_project_ids=["prj_x", "prj_y"],
            visibility="org",
        )
        meta_path = store._meta_path("meta_test", result.raw_id)
        assert meta_path.exists()
        meta = json.loads(meta_path.read_text(encoding="utf-8"))
        assert meta["raw_id"] == result.raw_id
        assert meta["source"] == "meta_test"
        assert meta["type"] == "dataset"
        assert meta["name"] == "users_2026"
        assert meta["registered_by"] == "u_owner"
        assert meta["owner_org_id"] == "ou_test_dept_a"
        assert meta["owner_project_ids"] == ["prj_x", "prj_y"]
        assert meta["visibility"] == "org"

    def test_load_populates_metadata_from_sidecar(
        self, temp_var_dir
    ) -> None:
        """load() should populate RawRecord fields from sidecar (not empty)."""
        store = get_raw_store()
        result = store.save(
            source="meta_test",
            type_="metric",
            name="qps",
            body="# QPS",
            registered_by="u_owner",
            owner_org_id="ou_test_dept_a",
        )
        record = store.load(source="meta_test", raw_id=result.raw_id)
        assert record.type == "metric"
        assert record.name == "qps"
        assert record.registered_by == "u_owner"
        assert record.owner_org_id == "ou_test_dept_a"
        assert record.visibility == "org"

    def test_delete_removes_meta_sidecar_too(
        self, temp_var_dir
    ) -> None:
        """delete() should remove both body + meta sidecar."""
        store = get_raw_store()
        result = store.save(source="meta_test", type_="dataset", name="x", body="x")
        meta_path = store._meta_path("meta_test", result.raw_id)
        assert meta_path.exists()
        deleted = store.delete(source="meta_test", raw_id=result.raw_id)
        assert deleted is True
        assert not meta_path.exists()
        assert not store.exists(source="meta_test", raw_id=result.raw_id)


class TestPerRawDEK:
    """Codex P2 review fix: generate per-raw DEK instead of reusing KEK."""

    def test_per_raw_dek_unique(self, temp_var_dir) -> None:
        """Two raws encrypted with same KEK must have different wrapped DEK."""
        kek = os.urandom(32)
        store = RawStore(base_dir=temp_var_dir / "raw", kek=kek)
        r1 = store.save(source="dek_test", type_="dataset", name="a", body="a")
        r2 = store.save(source="dek_test", type_="dataset", name="b", body="b")

        env1 = (temp_var_dir / "raw" / "dek_test" / f"{r1.raw_id}.bin").read_bytes()
        env2 = (temp_var_dir / "raw" / "dek_test" / f"{r2.raw_id}.bin").read_bytes()

        assert env1[0] == 2 and env2[0] == 2
        wrapped_dek_1 = env1[13:61]
        wrapped_dek_2 = env2[13:61]
        assert wrapped_dek_1 != wrapped_dek_2

    def test_envelope_v2_round_trip(self, temp_var_dir) -> None:
        """v2 envelope: encrypt + decrypt returns original plaintext."""
        kek = os.urandom(32)
        store = RawStore(base_dir=temp_var_dir / "raw", kek=kek)
        result = store.save(
            source="dek_test",
            type_="metric",
            name="cpu",
            body="cpu usage 80%",
        )
        record = store.load(source="dek_test", raw_id=result.raw_id)
        assert record.body == "cpu usage 80%"

    def test_wrong_kek_fails_dek_unwrap(self, temp_var_dir) -> None:
        """Wrong KEK fails to unwrap per-raw DEK (InvalidTag)."""
        kek_correct = os.urandom(32)
        kek_wrong = os.urandom(32)
        store_correct = RawStore(base_dir=temp_var_dir / "raw", kek=kek_correct)
        result = store_correct.save(
            source="dek_test", type_="dataset", name="x", body="secret"
        )
        store_wrong = RawStore(base_dir=temp_var_dir / "raw", kek=kek_wrong)
        with pytest.raises(Exception):
            store_wrong.load(source="dek_test", raw_id=result.raw_id)


class TestSourceNameValidation:
    """Codex P2 review fix: reject path traversal in source name."""

    @pytest.mark.parametrize("bad_name", [
        "../tmp",
        "/etc",
        "../../etc/passwd",
        "",
        "UPPERCASE",
        "with space",
        "with/slash",
        "with\\backslash",
        ".",
        "..",
    ])
    def test_invalid_source_name_rejected(
        self, temp_var_dir, bad_name: str
    ) -> None:
        """Path traversal / invalid source names raise InvalidSourceNameError."""
        store = get_raw_store()
        with pytest.raises(InvalidSourceNameError):
            store.save(source=bad_name, type_="dataset", name="x", body="x")

    def test_valid_source_name_accepted(self, temp_var_dir) -> None:
        """Valid source name (lowercase + hyphen + underscore) is accepted."""
        store = get_raw_store()
        result = store.save(
            source="valid-source_001", type_="dataset", name="x", body="x"
        )
        assert result.raw_id

    def test_source_name_pattern(self) -> None:
        """SOURCE_NAME_PATTERN matches valid names only."""
        assert SOURCE_NAME_PATTERN.match("homelab_mock")
        assert SOURCE_NAME_PATTERN.match("a")
        assert SOURCE_NAME_PATTERN.match("with-many-hyphens")
        assert SOURCE_NAME_PATTERN.match("with_underscores")
        assert not SOURCE_NAME_PATTERN.match("UPPER")
        assert not SOURCE_NAME_PATTERN.match("../bad")
        assert not SOURCE_NAME_PATTERN.match("/abs")
        assert not SOURCE_NAME_PATTERN.match("")
        assert not SOURCE_NAME_PATTERN.match("with space")
