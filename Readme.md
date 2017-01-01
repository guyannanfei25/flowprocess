Dispatcher
---
1.����
--
golang������Ϊ�˽���������⣬���ĵ���ʹ�����Ǹ����׽���������⣬���Ҽ����˽����˱���ż�������˱��Ч�ʡ����Ƕ��ڳ�ѧ�߻��Ǻ��ѺܺõĽ��в�����̡�
Ϊ�˸��õġ�����Ч�ı�̣�ʹ�ÿ�����Ա�ܹ�����Ĺ�ע��ҵ���߼��Ŀ����������˸���Ŀ���뷨��

2.���˼·
--
����Ŀ��������ʽ����ĳ�������һ��ͳһ�����dispatcher���ַ��������������⼶�����γ��û�����ͼ������ɲ�ͬ�����������Ϣ���У�������ʵ�ּ򵥵�storm��ʽ������Ŀ��

![framework.png](framework.png)

<center>(ʾ��ͼ)</center>

3.ʹ��
--
ʹ���߿��Լ̳���DefaultDispatcher������ʵ���Լ���Dispatcher�ӿڣ�����ǿ�ҽ���̳���DefaultDispatcher��Ȼ�������Ҫ�Զ��崦�����ˡ���Ҫ��ҵ���߼�������SetPreFunc��SetSufFuc��ʵ�֡�
�����ļ�����Ϊÿ��Dispatcher�ṩ�������������磺

    name        = framework
    concurrency = 3
    msgMaxSize  = 2

    [parse_dispatcher]
    name        = parse_dispatcher
    concurrency = 3
    msgMaxSize  = 2
```go
import (
    "github.com/guyannanfei25/flowprocess"
)

// ����ڿ���ֱ��ʹ��Ĭ��
var framework flowprocess.DefaultDispatcher

    framework.Init(ini, "")

    // �����û���ҵ��Dispatcher����DefaultDispatcher��Ͻ�����
    type ParseDispatcher struct {
        flowprocess.DefaultDispatcher
    }

    parse  := new(ParseDispatcher)
    parse.Init(ini, "parse_dispatcher")

    // �����û�ҵ������
    parse.SetPreFunc(parseline)

    // ����ͬҵ�������γ�����
    framework.DownRegister(parse)

    framework.Start()

    framework.Close()
```

logProcess����չʾ�򵥵���־�������̡�

4.ע������
---
����漰����Э�̰�ȫ�����⣬��Ҫʹ�����Լ���֤������Ŀ���ó���Ϊ��ʽ����ÿ��dispatcher������Ӧ��item���ݽ������Ρ���������ҵ������Ƕ����ģ��Ҳ�Ӧ��ȥ�޸�item����Ϊ����ͬ����Dispatcher�������item����˳�򲻹̶���

5.���ó���
---
������־����������item��һ����־������־��Ҫ���ݲ�ͬ��������з�������˺�������ʱ��Ϳ��Զ��Ʋ�ͬ�����dispatcher����Ҷ���������֮���һ�㼶��dispatcher�������ɷ�����������͸�������dispatcher��

6.TODO
---
���Ƴ�ʼ���ͽ����ķ���������:

1, ����DefaultDispatcher��Ա������Ϊ���������������������޸�DefaultDispatcher����ʼ���ͽ���������

2, ����DefaultDispatcher�ĳ�ʼ���ͽ������������Զ����ʼ����������֮��ȥ���á�

����1���������Ŀ����Զ��巽����ִ��˳�򣬵��ǶԴ�����������

����2����ԭʼ�����������ԣ��޷�������Ʋ������˳��
