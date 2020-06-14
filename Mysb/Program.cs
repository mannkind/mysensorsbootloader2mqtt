using System;
using System.Threading.Tasks;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;
using Mysb.DataAccess;
using Mysb.Liasons;
using TwoMQTT.Core;
using TwoMQTT.Core.Extensions;
using TwoMQTT.Core.Interfaces;
using TwoMQTT.Core.Utils;

namespace Mysb
{
    class Program : ConsoleProgram<object, object, SourceLiason, MQTTLiason>
    {
        static async Task Main(string[] args)
        {
            var p = new Program();
            await p.ExecuteAsync(args);
        }

        protected override IServiceCollection ConfigureServices(HostBuilderContext hostContext, IServiceCollection services)
        {
            return services
                .ConfigureOpts<Models.Options.SharedOpts>(hostContext, Models.Options.SharedOpts.Section)
                .ConfigureOpts<Models.Options.MQTTOpts>(hostContext, Models.Options.MQTTOpts.Section)
                .ConfigureOpts<TwoMQTT.Core.Models.MQTTManagerOptions>(hostContext, Models.Options.MQTTOpts.Section)
                .AddSingleton<IThrottleManager, ThrottleManager>(x =>
                {
                    return new ThrottleManager(new TimeSpan());
                })
                .AddSingleton<IFirmwareDAO, FirmwareDAO>(x =>
                {
                    var opts = x.GetService<IOptions<Models.Options.SharedOpts>>();
                    return new FirmwareDAO(x.GetService<ILogger<FirmwareDAO>>(),
                        opts.Value.FirmwareBasePath, opts.Value.Resources);
                });
        }
    }
}
