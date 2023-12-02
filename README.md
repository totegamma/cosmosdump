# cosmosdump

cosmosdump is a simple debug tool for Cosmos nodes.

When you write a non-deterministic code, such as using `time.Time()` or generating random values, you may encounter a consensus error like this:

```
ERR CONSENSUS FAILURE!!! err="+2/3 committed an invalid block: wrong Block.Header.AppHash.  Expected 5FC641EE89BB0579BE17943917CD5A7CD3C918E1C527247AFAD022C9755200C4, got 4F5E48398E8A0845DB8A7209FAFFD81C71F51F84A0E21C88059DB1C13650EF72"
```

By default, Cosmos provides no information about which store's differences are causing this AppHash mismatch, making debugging a challenging task.

cosmosdump addresses this issue. It's a simple script that dumps all key-value pairs from a Cosmos node's database. This allows you to compare and identify which keys are responsible for this discrepancy.

## usage

```
# install
go install

# dump
cosmosdump <dataDir> <height(optional)>
```

## Troubleshooting Scenario
Suppose you encounter a consensus error at block height 100. You can compare data from two different chains as follows:

```
$ cosmosdump ~/.my-chain1 100 > chain1.dump
$ cosmosdump ~/.my-chain2 100 > chain2.dump

$ diff chain1.dump chain2.dump -c1
*** dump1       2023-12-02 09:41:05.773711981 +0900
--- dump2       2023-12-02 09:41:09.943511826 +0900
***************
*** 1,3 ****
! dir: /home/totegamma/.my-chain1
! latestVersion:  175
  targetHeight:  101
--- 1,3 ----
! dir: /home/totegamma/.my-chain2
! latestVersion:  101
  targetHeight:  101
***************
*** 262,264 ****
  key(ascii): mymodule MyModule/value/hoge/
! value(hex): AAAAAA

--- 262,264 ----
  key(ascii): mymodule MyModule/value/hoge/
! value(hex): BBBBBB

Exit status: 1
[totegamma@09:41]~/tmp$
```

From this comparison, you can deduce that the key MyModule/value/hoge/ is the source of the issue.

In most cases, the values are protobuf-encoded. Therefore, you can inspect them using tools like [Protobuf Decoder](https://protobuf-decoder.netlify.app/)

