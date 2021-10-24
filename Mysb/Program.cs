using System;
using System.Threading.Tasks;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;
using Mysb.DataAccess;
using Mysb.Liasons;
using Mysb.Models.Options;
using TwoMQTT;
using TwoMQTT.Extensions;
using TwoMQTT.Interfaces;
using TwoMQTT.Managers;
using TwoMQTT.Utils;

await ConsoleProgram<object, object, SourceLiason, MQTTLiason>.ExecuteAsync(args,
    configureServices: (HostBuilderContext context, IServiceCollection services) =>
    {
        services
            .AddOptions<SharedOpts>(SharedOpts.Section, context.Configuration)
            .AddOptions<MQTTOpts>(MQTTOpts.Section, context.Configuration)
            .AddOptions<TwoMQTT.Models.MQTTManagerOptions>(MQTTOpts.Section, context.Configuration)
            .AddSingleton<IThrottleManager, ThrottleManager>(x =>
            {
                return new ThrottleManager(System.Threading.Timeout.InfiniteTimeSpan);
            })
            .AddSingleton<IFirmwareDAO, FirmwareDAO>(x =>
            {
                var logger = x.GetRequiredService<ILogger<FirmwareDAO>>();
                var opts = x.GetRequiredService<IOptions<SharedOpts>>();
                return new FirmwareDAO(logger, opts.Value.FirmwareBasePath, opts.Value.Resources);
            });
    });