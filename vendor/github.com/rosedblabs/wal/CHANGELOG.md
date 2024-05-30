# Release 1.3.6 (2023-09-25)

## 🎄 Enhancements
* avoid resetting pool to optimize the memory usage
* no need to return err in pendingWrites
* fix benchmark error
## 🎠 Community
* Thanks to @akiozihao 
    * check ErrPendingSizeTooLarge first (https://github.com/rosedblabs/wal/pull/32)

# Release 1.3.5 (2023-09-19)

## 🎄 Enhancements
* Rotate file when pending writes exceed the left space of the segment file.

# Release 1.3.4 (2023-09-18)

## 🚀 New Features
* add RenameFileExt function

## 🎠 Community
* Thanks to @akiozihao 
    * add EncodeFixedSize (https://github.com/rosedblabs/wal/pull/28)
    * add WriteBatch (https://github.com/rosedblabs/wal/pull/26)

# Release 1.3.3 (2023-08-19)

## 🎠 Community
* Thanks to @LEAVING-7 
  * Keep function name consistent in wal_test.go (https://github.com/rosedblabs/wal/pull/24)
* Thanks to @amityahav 
  * Improved performance for writing large records (> blockSize) (https://github.com/rosedblabs/wal/pull/21)
## 🐞 Bug Fixes
* fix a bug if the segment size exceeds 4GB
* Enhancement: use bufferpool to aviod writing twice https://github.com/rosedblabs/wal/commit/1345f5013113781c59ddaca36ddb13bdcc58ce27

# Release 1.3.2 (2023-08-07)

## 🎄 Enhancements
* Enhancement: use bufferpool to aviod writing twice https://github.com/rosedblabs/wal/commit/1345f5013113781c59ddaca36ddb13bdcc58ce27

# Release 1.3.1 (2023-08-04)

## 🐞 Bug Fixes
* Add a condition to avoid cache repeatedly https://github.com/rosedblabs/wal/commit/cb708139c877b1ef102c0be057ba33cb4af6abb2

# Release 1.3.0 (2023-08-02)

## 🚀 New Features
* Add ChunkPosition Encode and Decode

## 🎄 Enhancements
* Avoid to make new bytes while writing
* Use sync.Pool to optimize read performace
* Add more code comments

## 🎠 Community
* Thanks to @chinazmc 
  * update SementFileExt to SegmentFileExt (https://github.com/rosedblabs/wal/pull/11)
* Thanks to @xzhseh 
  * feat(docs): improve README.md format & fix several typos (https://github.com/rosedblabs/wal/pull/12)
* Thanks to @yanxiaoqi932 
  * BlockCache must smaller than SegmentSize (https://github.com/rosedblabs/wal/pull/14)
* Thanks to @mitingjin 
  * Fix typo in wal.go (https://github.com/rosedblabs/wal/pull/15)

# Release 1.2.0 (2023-07-01)

## 🚀 New Features
* Add `NewReaderWithStart` function to support read log from specified position.

## 🎠 Community
* Thanks to@yanxiaoqi932
  * enhancement: add wal delete function ([#7](https://github.com/rosedblabs/wal/pull/9))

# Release 1.1.0 (2023-06-21)

## 🚀 New Features
* Add tests in windows, with worlflow.
* Add some functions to support rosedb Merge operation.

## 🎠 Community
* Thanks to@SPCDTS
  * fix: calculate seg fle size by seg.size ([#7](https://github.com/rosedblabs/wal/pull/7))
  * fix: limit data size ([#6](https://github.com/rosedblabs/wal/pull/6))
  * fix: spelling error ([#5](https://github.com/rosedblabs/wal/pull/5))

# Release 1.0.0 (2023-06-13)

## 🚀 New Features
* First release, basic operations, read, write, and iterate the log files.
* Add block cache for log files.
