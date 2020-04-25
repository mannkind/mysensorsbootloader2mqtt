using Microsoft.VisualStudio.TestTools.UnitTesting;
using Mysb.Models.Shared;
using Mysb.DataAccess;
using System.Threading.Tasks;
using System.Collections.Generic;
using System.Threading;
using Microsoft.Extensions.Options;
using System.Linq;
using System.IO;
using Microsoft.Extensions.Logging;
using Moq;

namespace MysbTest
{
    [TestClass]
    public class FirmwareDAOTest
    {
        [TestMethod]
        public async Task FirmwareConfigAsyncTest()
        {
            var tests = new[]
            {
                new
                {
                    Payload = "010001005000D446",
                    Node = new NodeFirmwareInfoMapping { NodeId = string.Empty, Type = 1, Version = 1},
                    Expected = "010001005000D446"
                }
            };

            foreach (var test in tests)
            {
                var logger = new Mock<ILogger<ExposedFirmwareDAO>>().Object;
                var opts = Options.Create(new Opts
                {
                    Resources = new List<NodeFirmwareInfoMapping> { test.Node },
                    FirmwareBasePath = Const.TestFilesBasePath,
                });
                var dao = new ExposedFirmwareDAO(logger, opts);
                var actual = await dao.FirmwareConfigAsync(string.Empty, test.Payload);
                Assert.AreEqual(test.Expected, actual);
            }
        }

        [TestMethod]
        public async Task FirmwareAsyncTest()
        {
            var tests = new[]
            {
                new
                {
                    Payload = "010001000100",
                    Node = new NodeFirmwareInfoMapping { NodeId = string.Empty, Type = 1, Version = 1},
                    Expected = "0100010001000C946E000C946E000C946E000C946E00"
                },
                new
                {
                    Payload = "0B0001000100",
                    Node = new NodeFirmwareInfoMapping { NodeId = string.Empty, Type = 11, Version = 1},
                    Expected = "0B000100010052C1000050C10000EEC200004CC10000"
                }
            };

            foreach (var test in tests)
            {
                var logger = new Mock<ILogger<ExposedFirmwareDAO>>().Object;
                var opts = Options.Create(new Opts
                {
                    Resources = new List<NodeFirmwareInfoMapping> { test.Node },
                    FirmwareBasePath = Const.TestFilesBasePath,
                });
                var dao = new ExposedFirmwareDAO(logger, opts);
                var actual = await dao.FirmwareAsync(string.Empty, test.Payload);
                Assert.AreEqual(test.Expected, actual);
            }
        }

        [TestMethod]
        public async Task LoadFromFileAsyncTest()
        {
            var tests = new[]
            {
                new
                {
                    Path = $"{Const.TestFilesBasePath}/1/1/firmware.hex",
                    ExpectedBlocks = 80,
                    ExpectedCrc = 18132
                },
                new
                {
                    Path = $"{Const.TestFilesBasePath}/11/1/firmware.hex",
                    ExpectedBlocks = 1072,
                    ExpectedCrc = 64648
                }
            };

            foreach (var test in tests)
            {
                var logger = new Mock<ILogger<ExposedFirmwareDAO>>().Object;
                var opts = Options.Create(new Opts
                {
                    FirmwareBasePath = Const.TestFilesBasePath,
                });
                var dao = new ExposedFirmwareDAO(logger, opts);
                var firmware = await dao.LoadFromFileTestAsync(test.Path);

                Assert.AreEqual(test.ExpectedBlocks, firmware.Blocks);
                Assert.AreEqual(test.ExpectedCrc, firmware.Crc);
            }
        }

        [TestMethod]
        public async Task LoadFromFileContentAsyncTest()
        {
            var tests = new[]
            {
                new { Type = 1, Blocks = 80, Path = $"{Const.TestFilesBasePath}/1/1/firmware.encoded" },
                new { Type = 11, Blocks = 1072, Path = $"{Const.TestFilesBasePath}/11/1/firmware.encoded" }
            };

            foreach (var test in tests)
            {
                var logger = new Mock<ILogger<ExposedFirmwareDAO>>().Object;
                var opts = Options.Create(new Opts
                {
                    FirmwareBasePath = Const.TestFilesBasePath,
                });
                var dao = new ExposedFirmwareDAO(logger, opts);
                var lines = await File.ReadAllLinesAsync(test.Path);
                foreach (var blockNo in Enumerable.Range(0, test.Blocks))
                {
                    var lineNo = test.Blocks - blockNo - 1;
                    var payload = dao.PackTest(new FirmwareReqResp
                    {
                        Type = (ushort)test.Type,
                        Version = 1,
                        Block = (ushort)blockNo,
                    });

                    var actual = await dao.FirmwareAsync(string.Empty, payload);
                    var expected = lines[lineNo];

                    Assert.AreEqual(expected, actual, $"Type: {test.Type}, Line: {lineNo}, Block: {blockNo}");
                }
            }
        }

        [TestMethod]
        public void BootloaderTest()
        {
            var tests = new[]
            {
                new
                {
                    Description = "Erase EEPROM",
                    Topic = "mysensors/bootloader/1/1",
                    Payload = "",
                    Expected = "0100000000007ADA"
                },
                new
                {
                    Description = "Set NodeID",
                    Topic = "mysensors/bootloader/2/2",
                    Payload = "9",
                    Expected = "0200090000007ADA"
                },
               new
                {
                    Description = "Set ParentID",
                    Topic = "mysensors/bootloader/3/3",
                    Payload = "11",
                    Expected = "03000B0000007ADA"
                },
            };

            foreach (var test in tests)
            {
                var logger = new Mock<ILogger<ExposedFirmwareDAO>>().Object;
                var opts = Options.Create(new Opts
                {
                    FirmwareBasePath = Const.TestFilesBasePath,
                });
                var dao = new ExposedFirmwareDAO(logger, opts);
                var (_, actual) = dao.BootloaderCommand(test.Topic, test.Payload);
                Assert.AreEqual(test.Expected, actual, test.Description);
            }
        }
    }

    public class ExposedFirmwareDAO : FirmwareDAO
    {
        public ExposedFirmwareDAO(ILogger<ExposedFirmwareDAO> logger, IOptions<Opts> sharedOpts) : base(logger, sharedOpts)
        {
        }

        public Task<Firmware> LoadFromFileTestAsync(string path, CancellationToken cancellationToken = default) =>
            this.LoadFromFileAsync(path, cancellationToken);

        public string PackTest<T>(T obj) where T : struct => this.Pack(obj);
    }
}
