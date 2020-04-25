using System.Threading.Tasks;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.DependencyInjection;
using Mysb.Models.Shared;
using TwoMQTT.Core;
using Mysb.Managers;
using TwoMQTT.Core.Extensions;
using Microsoft.Extensions.Caching.Memory;
using Mysb.DataAccess;
using TwoMQTT.Core.DataAccess;

namespace Mysb
{
    class Program : ConsoleProgram
    {
        static async Task Main(string[] args)
        {
            var p = new Program();
            await p.ExecuteAsync(args);
        }

        protected override IServiceCollection ConfigureServices(HostBuilderContext hostContext, IServiceCollection services)
        {
            var sinkSect = hostContext.Configuration.GetSection(Models.SinkManager.Opts.Section);
            var sharedSect = hostContext.Configuration.GetSection(Models.Shared.Opts.Section);

            return services
                .Configure<Models.SinkManager.Opts>(sinkSect)
                .Configure<Models.Shared.Opts>(sharedSect)
                .AddSingleton<IFirmwareDAO, FirmwareDAO>()
                .AddSingleton<IHostedService, SinkManager>();
        }
    }
}