package flow

const (
	DotTpl = `digraph en_name_audit {
    bgcolor="lightyellow"
    rankdir="LR"
    #布局第一列
    {rank="same";}

    graph[label="测试dot" comment="{\"start_event\":\"event_start\",\"version\":\"v0.0.1\",\"no_strict\":true,\"remark\":\"测试dot备注\"}"]
    {
        #事件
        event_start [shape=circle label="开始" style=filled fillcolor=black fontcolor=green];
        event_end_flow [shape=circle label="结束" style=filled fillcolor=black fontcolor=red];

        #活动
        activity_open_first [shape=box label="先发后审" style=filled fillcolor=gold];
        activity_audit_first [shape=box label="先审后发" style=filled fillcolor=gold];
		activity_pay_audit [shape=box label="付费审核" style=filled fillcolor=gold];

        #网关
        gateway_auto_distribution [shape=diamond label="自动分发网关" comment="{\"type\":\"xor\"}" style=filled fillcolor=pink];
		gateway_pay_audit [shape=diamond label="付费审核网关" comment="{\"type\":\"xor\"}" style=filled fillcolor=pink];
    }

   	# 开始=>自动分发网关
	event_start->gateway_auto_distribution;

	# 自动分发网关=>先发后审、先审后发、付费审核
    gateway_auto_distribution->activity_pay_audit [label="命中付费审核" comment="{\"rule\":\"metadata.pay_season == 1\",\"sort\":3}"];
	gateway_auto_distribution->activity_open_first [label="命中先发后审" comment="{\"rule\":\"extra6 == 1\",\"sort\":2}"];
	gateway_auto_distribution->activity_audit_first [label="命中先审后发" comment="{\"rule\":\"extra6 != 1\",\"sort\":1}"];

	# 付费审核=>付费审核网关
	activity_pay_audit->gateway_pay_audit;

	# 付费审核网关=>先审后发、结束
	gateway_pay_audit->activity_audit_first [label="命中先审后发" comment="{\"rule\":\"state == 0\",\"sort\":2}"];
	gateway_pay_audit->event_end_flow [label="命中结束" comment="{\"rule\":\"state == -2 || state == -4\",\"sort\":1}"];

	# 先发后审=>结束
    activity_open_first->event_end_flow;

	# 先审后发=>结束
	activity_audit_first->event_end_flow;
}`
)
