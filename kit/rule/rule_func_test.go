package rule


//{
//name: "Single function",
//expr: "foo()",
//extension: Function("foo", func(arguments ...interface{}) (interface{}, error) {
//return true, nil
//}),
//
//want: true,
//},
//{
//name: "Func with argument",
//expr: "passthrough(1)",
//extension: Function("passthrough", func(arguments ...interface{}) (interface{}, error) {
//return arguments[0], nil
//}),
//want: 1.0,
//},
//{
//name: "Func with arguments",
//expr: "passthrough(1, 2)",
//extension: Function("passthrough", func(arguments ...interface{}) (interface{}, error) {
//return arguments[0].(float64) + arguments[1].(float64), nil
//}),
//want: 3.0,
//},
//{
//name: "Nested function with operatorPrecedence",
//expr: "sum(1, sum(2, 3), 2 + 2, true ? 4 : 5)",
//extension: Function("sum", func(arguments ...interface{}) (interface{}, error) {
//sum := 0.0
//for _, v := range arguments {
//sum += v.(float64)
//}
//return sum, nil
//}),
//want: 14.0,
//},
//{
//name: "Empty function and modifier, compared",
//expr: "numeric()-1 > 0",
//extension: Function("numeric", func(arguments ...interface{}) (interface{}, error) {
//return 2.0, nil
//}),
//want: true,
//},
//{
//name: "Empty function comparator",
//expr: "numeric() > 0",
//extension: Function("numeric", func(arguments ...interface{}) (interface{}, error) {
//return 2.0, nil
//}),
//want: true,
//},
//{
//
//name: "Empty function logical operator",
//expr: "success() && !false",
//extension: Function("success", func(arguments ...interface{}) (interface{}, error) {
//return true, nil
//}),
//want: true,
//},
//{
//name: "Empty function ternary",
//expr: "nope() ? 1 : 2.0",
//extension: Function("nope", func(arguments ...interface{}) (interface{}, error) {
//return false, nil
//}),
//want: 2.0,
//},
//{
//
//name: "Empty function null coalesce",
//expr: "null() ?? 2",
//extension: Function("null", func(arguments ...interface{}) (interface{}, error) {
//return nil, nil
//}),
//want: 2.0,
//},
//{
//name: "Empty function with prefix",
//expr: "-ten()",
//extension: Function("ten", func(arguments ...interface{}) (interface{}, error) {
//return 10.0, nil
//}),
//want: -10.0,
//},
//{
//name: "Empty function as part of chain",
//expr: "10 - numeric() - 2",
//extension: Function("numeric", func(arguments ...interface{}) (interface{}, error) {
//return 5.0, nil
//}),
//want: 3.0,
//},
//{
//name: "Empty function near separator",
//expr: "10 in [1, 2, 3, ten(), 8]",
//extension: Function("ten", func(arguments ...interface{}) (interface{}, error) {
//return 10.0, nil
//}),
//want: true,
//},
//{
//name: "Enclosed empty function with modifier and comparator (#28)",
//expr: "(ten() - 1) > 3",
//extension: Function("ten", func(arguments ...interface{}) (interface{}, error) {
//return 10.0, nil
//}),
//want: true,
//},
//{
//name: "Array",
//expr: `[(ten() - 1) > 3, (ten() - 1),"hey"]`,
//extension: Function("ten", func(arguments ...interface{}) (interface{}, error) {
//return 10.0, nil
//}),
//want: []interface{}{true, 9., "hey"},
//},
//{
//name: "Object",
//expr: `{1: (ten() - 1) > 3, 7 + ".X" : (ten() - 1),"hello" : "hey"}`,
//extension: Function("ten", func(arguments ...interface{}) (interface{}, error) {
//return 10.0, nil
//}),
//want: map[string]interface{}{"1": true, "7.X": 9., "hello": "hey"},
//},
//{
//name: "Object negativ value",
//expr: `{1: -1,"hello" : "hey"}`,
//want: map[string]interface{}{"1": -1., "hello": "hey"},
//},
//{
//name: "Empty Array",
//expr: `[]`,
//want: []interface{}{},
//},
//{
//name: "Empty Object",
//expr: `{}`,
//want: map[string]interface{}{},
//},
//{
//name: "Variadic",
//expr: `sum(1,2,3,4)`,
//extension: Function("sum", func(arguments ...float64) (interface{}, error) {
//sum := 0.
//for _, a := range arguments {
//sum += a
//}
//return sum, nil
//}),
//want: 10.0,
//},
//{
//name: "Ident Operator",
//expr: `1 plus 1`,
//extension: InfixNumberOperator("plus", func(a, b float64) (interface{}, error) {
//return a + b, nil
//}),
//want: 2.0,
//},
//{
//name: "Postfix Operator",
//expr: `4§`,
//extension: PostfixOperator("§", func(_ context.Context, _ *Parser, eval Evaluable) (Evaluable, error) {
//return func(ctx context.Context, parameter interface{}) (interface{}, error) {
//i, err := eval.EvalInt(ctx, parameter)
//if err != nil {
//return nil, err
//}
//return fmt.Sprintf("§%d", i), nil
//}, nil
//}),
//want: "§4",
//},
//{
//name: "Tabs as non-whitespace",
//expr: "4\t5\t6",
//extension: NewLanguage(
//Init(func(ctx context.Context, p *Parser) (Evaluable, error) {
//p.SetWhitespace('\n', '\r', ' ')
//return p.ParseExpression(ctx)
//}),
//InfixNumberOperator("\t", func(a, b float64) (interface{}, error) {
//return a * b, nil
//}),
//),
//want: 120.0,
//},
//{
//name: "Handle all other prefixes",
//expr: "^foo + $bar + &baz",
//extension: DefaultExtension(func(ctx context.Context, p *Parser) (Evaluable, error) {
//var mul int
//switch p.TokenText() {
//case "^":
//mul = 1
//case "$":
//mul = 2
//case "&":
//mul = 3
//}
//
//switch p.Scan() {
//case scanner.Ident:
//return p.Const(mul * len(p.TokenText())), nil
//default:
//return nil, p.Expected("length multiplier", scanner.Ident)
//}
//}),
//want: 18.0,
//},
//{
//name: "Embed languages",
//expr: "left { 5 + 5 } right",
//extension: func() Language {
//step := func(ctx context.Context, p *Parser, cur Evaluable) (Evaluable, error) {
//next, err := p.ParseExpression(ctx)
//if err != nil {
//return nil, err
//}
//
//return func(ctx context.Context, parameter interface{}) (interface{}, error) {
//us, err := cur.EvalString(ctx, parameter)
//if err != nil {
//return nil, err
//}
//
//them, err := next.EvalString(ctx, parameter)
//if err != nil {
//return nil, err
//}
//
//return us + them, nil
//}, nil
//}
//
//return NewLanguage(
//Init(func(ctx context.Context, p *Parser) (Evaluable, error) {
//p.SetWhitespace()
//p.SetMode(0)
//
//return p.ParseExpression(ctx)
//}),
//DefaultExtension(func(ctx context.Context, p *Parser) (Evaluable, error) {
//return step(ctx, p, p.Const(p.TokenText()))
//}),
//PrefixExtension(scanner.EOF, func(ctx context.Context, p *Parser) (Evaluable, error) {
//return p.Const(""), nil
//}),
//PrefixExtension('{', func(ctx context.Context, p *Parser) (Evaluable, error) {
//eval, err := p.ParseSublanguage(ctx, Full())
//if err != nil {
//return nil, err
//}
//
//switch p.Scan() {
//case '}':
//default:
//return nil, p.Expected("embedded", '}')
//}
//
//return step(ctx, p, eval)
//}),
//)
//}(),
//want: "left 10 right",
//},
//{
//name: "Late binding",
//expr: "5 * [ 10 * { 20 / [ 10 ] } ]",
//extension: func() Language {
//var inner, outer Language
//
//parseCurly := func(ctx context.Context, p *Parser) (Evaluable, error) {
//eval, err := p.ParseSublanguage(ctx, outer)
//if err != nil {
//return nil, err
//}
//
//if p.Scan() != '}' {
//return nil, p.Expected("end", '}')
//}
//
//return eval, nil
//}
//
//parseSquare := func(ctx context.Context, p *Parser) (Evaluable, error) {
//eval, err := p.ParseSublanguage(ctx, inner)
//if err != nil {
//return nil, err
//}
//
//if p.Scan() != ']' {
//return nil, p.Expected("end", ']')
//}
//
//return eval, nil
//}
//
//inner = Full(PrefixExtension('{', parseCurly))
//outer = Full(PrefixExtension('[', parseSquare))
//return outer
//}(),
//want: 100.0,
//},