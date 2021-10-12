using System;
using System.Threading.Tasks;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;
using Mysb.DataAccess;
using Mysb.Liasons;
using TwoMQTT;
using TwoMQTT.Extensions;
using TwoMQTT.Interfaces;
using TwoMQTT.Managers;
using TwoMQTT.Utils;

namespace Mysb
{
    class Program
    {
        static async Task Main(string[] args) =>
            await ConsoleProgram<object, object, SourceLiason, MQTTLiason>.ExecuteAsync(args,
                configureServices: (HostBuilderContext context, IServiceCollection services) =>
                {
                    services
                        .AddOptions<Models.Options.SharedOpts>(Models.Options.SharedOpts.Section, context.Configuration)
                        .AddOptions<Models.Options.MQTTOpts>(Models.Options.MQTTOpts.Section, context.Configuration)
                        .AddOptions<TwoMQTT.Models.MQTTManagerOptions>(Models.Options.MQTTOpts.Section, context.Configuration)
                        .AddSingleton<IThrottleManager, ThrottleManager>(x =>
                        {
                            return new ThrottleManager(System.Threading.Timeout.InfiniteTimeSpan);
                        })
                        .AddSingleton<IFirmwareDAO, FirmwareDAO>(x =>
                        {
                            var logger = x.GetRequiredService<ILogger<FirmwareDAO>>();
                            var opts = x.GetRequiredService<IOptions<Models.Options.SharedOpts>>();
                            return new FirmwareDAO(logger, opts.Value.FirmwareBasePath, opts.Value.Resources);
                        });
                });
    }
}
