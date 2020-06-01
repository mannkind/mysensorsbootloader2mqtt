using System.Collections.Concurrent;
using System.Collections.Generic;
using System.Threading;
using System.Threading.Channels;
using System.Threading.Tasks;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;
using MQTTnet;
using Mysb.DataAccess;
using Mysb.Models.Shared;
using TwoMQTT.Core.Managers;

namespace Mysb.Managers
{
    /// <summary>
    /// An class representing a managed way to interact with a sink.
    /// </summary>
    public class SinkManager : MQTTManager<NodeFirmwareInfoMapping, object, object>
    {
        /// <summary>
        /// Initializes a new instance of the SinkManager class.
        /// </summary>
        /// <param name="logger"></param>
        /// <param name="sharedOpts"></param>
        /// <param name="opts"></param>
        /// <param name="incomingData"></param>
        /// <param name="outgoingCommand"></param>
        /// <param name="loader"></param>
        /// <typeparam name="object"></typeparam>
        /// <returns></returns>
        public SinkManager(ILogger<SinkManager> logger, IOptions<Opts> sharedOpts,
            IOptions<Models.SinkManager.Opts> opts, ChannelReader<object> incomingData, ChannelWriter<object> outgoingCommand, IFirmwareDAO loader) :
            base(logger, opts, incomingData, outgoingCommand, sharedOpts.Value.Resources, SinkSettings(sharedOpts.Value))
        {
            this.AutoIDEnabled = sharedOpts.Value.AutoIDEnabled;
            this.NextID = sharedOpts.Value.NextID;
            this.SubTopic = sharedOpts.Value.SubTopic;
            this.PubTopic = sharedOpts.Value.PubTopic;
            this.FirmwareDAO = loader;
        }

        /// <inheritdoc />
        protected override async Task HandleSubscribeAsync(CancellationToken cancellationToken = default)
        {
            var topics = new List<string>
            {
                $"{this.SubTopic}/{Const.IdRequestTopicPartial}",
                $"{this.SubTopic}/{Const.FirmwareConfigRequestTopicPartial}",
                $"{this.SubTopic}/{Const.FirmwareRequestTopicPartial}",
                Const.FirmwareBootloaderCommandTopic,
            };

            await this.SubscribeAsync(topics, cancellationToken);
        }

        /// <inheritdoc />
        protected override async Task HandleIncomingMessageAsync(string topic, string payload,
            CancellationToken cancellationToken = default)
        {
            // await base.HandleIncomingMessageAsync(topic, payload, cancellationToken);
            var bootloaderCommand = Const.FirmwareBootloaderCommandTopic.Replace("/+/+", string.Empty);
            if (topic.StartsWith(bootloaderCommand))
            {
                this.HandleBootloaderCommand(topic, payload, cancellationToken);
                return;
            }

            var parts = topic.Replace($"{this.SubTopic}/", string.Empty).Split('/');
            if (parts.Length != 5)
            {
                this.Logger.LogError("Unable to determine the nodeId from the topic; aborting.");
                return;
            }

            var nodeId = parts[0];
            var idRequest = $"{this.SubTopic}/{Const.IdRequestTopicPartial}";
            var firmwareConfigRequest = $"{this.SubTopic}/{nodeId}/{Const.FirmwareConfigRequestTopicPartial}".Replace("+/", string.Empty);
            var firmwareRequest = $"{this.SubTopic}/{nodeId}/{Const.FirmwareRequestTopicPartial}".Replace("+/", string.Empty);

            switch (topic)
            {
                case string s when s == idRequest:
                    await this.HandleIdRequest(cancellationToken);
                    break;

                case string s when s == firmwareConfigRequest:
                    await this.HandleFirmwareConfigRequest(nodeId, payload, cancellationToken);
                    break;

                case string s when s == firmwareRequest:
                    await this.HandleFirmwareRequest(nodeId, payload, cancellationToken);
                    break;

                case string s when s.StartsWith(bootloaderCommand):
                    break;
            }
        }

        /// <summary>
        /// Publish topics + payloads
        /// </summary>
        protected new async Task PublishAsync(string topic, string payload, CancellationToken cancellationToken = default)
        {
            this.Logger.LogDebug($"Publishing '{payload}' on '{topic}'");
            await this.Client.PublishAsync(
                new MqttApplicationMessageBuilder()
                    .WithTopic(topic)
                    .WithPayload(payload)
                    .WithExactlyOnceQoS()
                    .Build(),
                cancellationToken
            );
        }

        /// <inheritdoc />
        protected override Task HandleIncomingDataAsync(object input, CancellationToken cancellationToken = default) =>
            Task.CompletedTask;

        /// <inheritdoc />
        protected override Task HandleDiscoveryAsync(CancellationToken cancellationToken = default) =>
            Task.CompletedTask;

        /// <summary>
        /// 
        /// </summary>
        private ushort NextID;

        /// <summary>
        /// 
        /// </summary>
        private readonly bool AutoIDEnabled;

        /// <summary>
        /// 
        /// </summary>
        private readonly string SubTopic;

        /// <summary>
        /// 
        /// </summary>
        private readonly string PubTopic;

        /// <summary>
        /// 
        /// </summary>
        private readonly IFirmwareDAO FirmwareDAO;

        /// <summary>
        /// 
        /// </summary>
        /// <typeparam name="string"></typeparam>
        /// <typeparam name="FirmwareConfigReqResp"></typeparam>
        /// <returns></returns>
        private readonly ConcurrentDictionary<string, string> BootloaderCommands = new ConcurrentDictionary<string, string>();

        /// <summary>
        /// 
        /// </summary>
        /// <param name="cancellationToken"></param>
        /// <returns></returns>
        private async Task HandleIdRequest(CancellationToken cancellationToken = default)
        {
            if (!this.AutoIDEnabled)
            {
                return;
            }

            this.NextID += 1;

            var respTopic = $"{this.PubTopic}/{Const.IdResponseTopicPartial}";
            var respPayload = this.NextID.ToString();
            await this.PublishAsync(respTopic, respPayload, cancellationToken);
        }

        /// <summary>
        /// 
        /// </summary>
        /// <param name="nodeId"></param>
        /// <param name="payload"></param>
        /// <param name="cancellationToken"></param>
        /// <returns></returns>
        private async Task HandleFirmwareConfigRequest(string nodeId, string payload, CancellationToken cancellationToken = default)
        {
            var respTopic = $"{this.PubTopic}/{nodeId}/{Const.FirmwareConfigResponseTopicPartial}";

            if (this.BootloaderCommands.ContainsKey(nodeId) && this.BootloaderCommands.TryRemove(nodeId, out var bootloaderPayload))
            {
                await this.PublishAsync(respTopic, bootloaderPayload, cancellationToken);
                return;
            }

            var respPayload = await this.FirmwareDAO.FirmwareConfigAsync(nodeId, payload, cancellationToken);
            await this.PublishAsync(respTopic, respPayload, cancellationToken);
        }

        /// <summary>
        /// 
        /// </summary>
        /// <param name="nodeId"></param>
        /// <param name="payload"></param>
        /// <param name="cancellationToken"></param>
        /// <returns></returns>
        private async Task HandleFirmwareRequest(string nodeId, string payload, CancellationToken cancellationToken = default)
        {
            var respTopic = $"{this.PubTopic}/{nodeId}/{Const.FirmwareResponseTopicPartial}";
            var respPayload = await this.FirmwareDAO.FirmwareAsync(nodeId, payload, cancellationToken);
            await this.PublishAsync(respTopic, respPayload, cancellationToken);
        }

        /// <summary>
        /// 
        /// </summary>
        /// <param name="topic"></param>
        /// <param name="payload"></param>
        /// <param name="cancellationToken"></param>
        /// <returns></returns>
        private void HandleBootloaderCommand(string topic, string payload, CancellationToken cancellationToken = default)
        {
            var (nodeId, resp) = this.FirmwareDAO.BootloaderCommand(topic, payload);
            if (string.IsNullOrEmpty(nodeId))
            {
                return;
            }

            this.BootloaderCommands[nodeId] = resp;
            return;
        }

        private static string SinkSettings(Models.Shared.Opts sharedOpts) =>
            $"FirmwareBasePath: {sharedOpts.FirmwareBasePath}\n" +
            $"AutoIDEnabled: {sharedOpts.AutoIDEnabled}\n" +
            $"NextID: {sharedOpts.NextID}\n" +
            $"SubTopic: {sharedOpts.SubTopic}\n" +
            $"PubTopic: {sharedOpts.PubTopic}\n" +
            $"Resources: {string.Join("; ", sharedOpts.Resources)}\n" +
            $"";
    }
}