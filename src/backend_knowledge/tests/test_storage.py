"""Raw storage test (umbrella doc §3.6 / ADR-0025 봉투 암호화)."""

from __future__ import annotations

import pytest

from backend_knowledge.storage import RawStore, get_raw_store


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
        assert record.registered_by == ""

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
        import os
        with pytest.raises(Exception, match="KEK must be 32 bytes"):
            RawStore(base_dir=temp_var_dir / "raw", kek=os.urandom(16))
