#include <stdio.h>
#include <unistd.h>
#include <stdlib.h>
#include <string.h>
#include <fcntl.h>
#include <pthread.h>
#include <arpa/inet.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <sys/epoll.h>
#include <pthread.h>
#include <sys/resource.h>

#define ADDR_LEN 16
#define MAX_PATH 256
#define EPOLL_SIZE 1000
#define CONN_TIMEOUT 3

typedef struct {
    int port;
    char addr[ADDR_LEN];

    int sockfd;
    int result;
} ScanPort;

typedef struct {
    int scan_method;
    int start_port;
    int end_port;
    char addr[ADDR_LEN];
    int threadcount;
    char failPath[MAX_PATH];
    char succPath[MAX_PATH];
    int limit;
    int epfd;
    int timeout;
    ScanPort *scan;
    FILE* succ;
    FILE* fail;
    int bExit;
    int bScan;
    int bWait;
    pthread_mutex_t *scan_mtx;
    pthread_mutex_t *wait_mtx;
    pthread_cond_t *scan_cond;
    pthread_cond_t *wait_cond;
} ScanParam;

static void printHelp(char *prog);
static void initScan(ScanPort* scan, int count);
static void parsePort(ScanParam *param, char *str);

static void *do_scanport(void *arg);
static void *wait_scanport(void *arg);
static void write_scan_file(FILE* fp, ScanPort *scan);

static void do_scanport_limit(ScanParam *param, int limit);
static void initScanPort(ScanPort *scan, char *addr, int port);

//static void freeScan(ScanPort* scan, int count);
//static void write_result_file(ScanParam *param);

int main(int argc,char *argv[])
{
    int epfd;
    int count;
    int ch = 0;

    ScanPort *scan;
    ScanParam param;

    int ret;
    int *ret_join = NULL;
    pthread_t scan_pt;
    pthread_t wait_pt;

    struct rlimit rlim;

    pthread_mutex_t scan_mtx;
    pthread_mutex_t wait_mtx;
    pthread_cond_t scan_cond;
    pthread_cond_t wait_cond;

    memset(&param, 0, sizeof(ScanParam));

    pthread_mutex_init(&scan_mtx, NULL);
    pthread_mutex_init(&wait_mtx, NULL);
    pthread_cond_init(&scan_cond, NULL);
    pthread_cond_init(&wait_cond, NULL);

    param.bExit = 0;
    param.bScan = 0;
    param.bWait = 0;

    param.scan_mtx = &scan_mtx;
    param.wait_mtx = &wait_mtx;
    param.scan_cond = &scan_cond;
    param.wait_cond = &wait_cond;
    
    //printf("command args=%d\n", argc);
    if (argc == 2)
    {
        if (!strcmp(argv[1], "?") ||
            !strcmp(argv[1], "--help"))
        {
            printHelp(argv[0]);
            exit(0);
        }
    }
    else if (argc < 15)
    {
        printHelp(argv[0]);
        exit(0);
    }
    
    while ((ch = getopt(argc, argv, "s:p:h:l:t:f:S:")) != -1)
    {
        switch (ch)
        {
        case 's':
            //printf("s args=%s\n", optarg);
            if (strcasecmp(optarg, "tcp") == 0)
            {
                param.scan_method = 1;
            }
            else
            {
                param.scan_method = 0;
            }
            break;
        case 'p':
            //printf("p args=%s\n", optarg);
            parsePort(&param, optarg);
            break;
        case 'h':
            //printf("h args=%s\n", optarg);
            strcpy(param.addr, optarg);
            break;
        case 'l':
            //printf("t args=%s\n", optarg);
            param.limit = atoi(optarg);
            break;
        case 't':
            //printf("t args=%s\n", optarg);
            param.timeout = atoi(optarg);
            break;
       case 'f':
            //printf("f args=%s\n", optarg);
            strcpy(param.failPath, optarg);
            break;
       case 'S':
            //printf("S args=%s\n", optarg);
            strcpy(param.succPath, optarg);
            break;
        default:
            //printHelp(argv[0]);
            exit(0);
            break;
        }
    }

    if (param.timeout == 0) {
        param.timeout = CONN_TIMEOUT;
    }

    if (param.limit == 0) {
        param.limit = EPOLL_SIZE;
    }
    if (param.limit <= 0) {
        printf("max sockets limit is invailid.\n");

        pthread_mutex_destroy(&scan_mtx);
        pthread_mutex_destroy(&wait_mtx);

        return 0;
    }

    if (getrlimit(RLIMIT_NOFILE, &rlim) != 0)
    {
        printf("get sockets limit fail.\n");

        pthread_mutex_destroy(&scan_mtx);
        pthread_mutex_destroy(&wait_mtx);

        return 0;
    }

    rlim.rlim_cur -= 50;
    if (rlim.rlim_cur <= 0)
    {
        printf("current sockets limit is invalid.\n");

        pthread_mutex_destroy(&scan_mtx);
        pthread_mutex_destroy(&wait_mtx);

        return 0;
    }

    //printf("get sockets limit val=%d.\n", rlim.rlim_cur);

    if (param.limit > rlim.rlim_cur)
    {
        param.limit = rlim.rlim_cur;
    }

    count = param.limit;
    scan = (ScanPort*)malloc(count * sizeof(ScanPort));
    if (scan == NULL)  {

        pthread_mutex_destroy(&scan_mtx);
        pthread_mutex_destroy(&wait_mtx);

        return 0;
    }
    param.scan = scan;

    epfd = epoll_create(count);
    if (epfd == -1)  {
        free(scan);

        pthread_mutex_destroy(&scan_mtx);
        pthread_mutex_destroy(&wait_mtx);

        return 0;
    }
    param.epfd = epfd;

    param.fail = fopen(param.failPath, "wb");
    if (param.fail == NULL) {
        free(scan);
        close(epfd);

        pthread_mutex_destroy(&scan_mtx);
        pthread_mutex_destroy(&wait_mtx);

        return 0;
    }

    param.succ = fopen(param.succPath, "wb");
    if (param.succ == NULL) {
        free(scan);
        close(epfd);
        fclose(param.fail);

        pthread_mutex_destroy(&scan_mtx);
        pthread_mutex_destroy(&wait_mtx);

        return 0;
    }

    ret = pthread_create(&scan_pt, NULL, (void*)do_scanport, &param); 
    if (ret != 0)
    {
        free(scan);
        close(epfd);
        fclose(param.fail);
        fclose(param.succ);

        pthread_mutex_destroy(&scan_mtx);
        pthread_mutex_destroy(&wait_mtx);

        return 1;
    }

    ret = pthread_create(&wait_pt, NULL, (void*)wait_scanport, &param); 
    if (ret != 0)
    {
        free(scan);
        close(epfd);
        fclose(param.fail);
        fclose(param.succ);

        pthread_mutex_destroy(&scan_mtx);
        pthread_mutex_destroy(&wait_mtx);

        return 1;
    }

    pthread_join(scan_pt, (void*)&ret_join);
    pthread_join(wait_pt, (void*)&ret_join); 

    free(scan);
    close(epfd);
    fclose(param.succ);
    fclose(param.fail);

    pthread_mutex_destroy(&scan_mtx);
    pthread_mutex_destroy(&wait_mtx);

    printf("program scan finish.\n");

    return 0;
}

static void write_scan_file(FILE* fp, ScanPort *scan)
{
    size_t size;
    char buf[128];
    memset(buf, 0, 128);

    sprintf(buf, "%s:%d\n", scan->addr, scan->port); 
    size = strlen(buf);

    fseek(fp, 0, SEEK_END);
    fwrite(buf, (size_t)size, 1, fp);
}

static void do_scanport_limit(ScanParam *param, int limit)
{
    int epfd;
    int port;
    char *addr;
    int sockfd;  

    int k;
    ScanPort *p;
    ScanPort *scan;

    struct epoll_event ev;
    struct sockaddr_in saddr;

    epfd = param->epfd;
    scan = param->scan;
    
    for (k = 0; k < limit; k++)
    {
        p = scan + k;

        port = p->port;
        addr = p->addr;
        sockfd = p->sockfd;
        if (sockfd == -1) continue;

        saddr.sin_family = AF_INET;
        saddr.sin_port = htons(port);
        saddr.sin_addr.s_addr = inet_addr(addr);

        if (connect(sockfd, (struct sockaddr*)&saddr, sizeof(struct sockaddr)) == 0) 
        {
            p->result = 1;
            printf("connect succ Port = %d\n", port);
            continue;
        }
        
        ev.data.ptr = p;
        ev.events = EPOLLOUT|EPOLLIN;

        if (epoll_ctl(epfd, EPOLL_CTL_ADD, sockfd, &ev) == -1) 
        {
            printf("connect epoll_ctl error\n");
            return;
        }
    }

    //printf("do_scanport begin scan\n");
    pthread_mutex_lock(param->scan_mtx);
    param->bExit = 0;
    param->bScan = 1;
    pthread_cond_signal(param->scan_cond);
    pthread_mutex_unlock(param->scan_mtx);           

    pthread_mutex_lock(param->wait_mtx);
    while (param->bWait == 0)
    {
        //printf("do_scanport wait scan\n");
        pthread_cond_wait(param->wait_cond, param->wait_mtx);
    }

    param->bWait = 0;
    //printf("do_scanport wait scan return\n");
    for (k = 0; k < limit; k++)
    {
        p = scan + k;
        sockfd = p->sockfd;

        if (sockfd == -1) 
        {
            printf("do_scanport socket invalid port=%d", p->port);
            continue;
        }
        
        if (p->result == 1)
        {
            write_scan_file(param->succ, p);
        }
        else
        {
            write_scan_file(param->fail, p);
        }

        ev.data.ptr = p;
        ev.events = EPOLLOUT|EPOLLIN;
        epoll_ctl(epfd, EPOLL_CTL_DEL, sockfd, NULL);

        close(sockfd);
    }
    pthread_mutex_unlock(param->wait_mtx);
}

static void initScanPort(ScanPort *scan, char *addr, int port)
{
    int sockfd;  
    int flags;  
   
    scan->port = port;
    strcpy(scan->addr, addr);

    sockfd = socket(AF_INET, SOCK_STREAM, 0);
    if (sockfd < 0)
    {
        printf("do_scanport ceate socket fail fd=%d\n", sockfd);
        return;
    }
    
    flags = fcntl(sockfd, F_GETFL, 0);
    fcntl(sockfd, F_SETFL, flags | O_NONBLOCK);

    scan->result = 0;
    scan->sockfd = sockfd;
}

static void *do_scanport(void *arg)
{
    int epfd;
    int limit;
    int count;
    ScanParam *param;

    ScanPort *p;
    ScanPort *scan;

    int port;
    char *addr;

    int i, j;

    param = (ScanParam *)arg;
    epfd = param->epfd;
    limit = param->limit;
    scan = param->scan;

    count = param->end_port - param->start_port + 1;

    //printf("do_scanport count=%d\n", count);

    j = 0;
    initScan(scan, limit);

    for (i = 0; i < count; i++)
    {
        port = param->start_port + i;
        addr = param->addr;

        if (j < limit)
        {
            p = scan + j;
            initScanPort(p, addr, port);
            j++;
        }
        else
        {
            //printf("do_scanport limit=%d\n", limit);

            do_scanport_limit(param, limit);

            j = 0;
            initScan(scan, limit);

             p = scan + j;
            initScanPort(p, addr, port);
            j++;
        }
    }

    if (j > 0)
    {
        //printf("do_scanport j=%d\n", j);

        do_scanport_limit(param, j);
    }

    //printf("do_scanport begin scan exit\n");
    pthread_mutex_lock(param->scan_mtx);
    param->bExit = 1;
    param->bScan = 0;
    pthread_cond_signal(param->scan_cond);
    pthread_mutex_unlock(param->scan_mtx);
}

static void *wait_scanport(void *arg)
{
    int epfd; 
    int limit;
    int timeout;
    ScanParam *param;

    int i;
    int nfds;
    int sockfd;

    ScanPort *scan;
    struct epoll_event ev;
    struct epoll_event *events;

    int error;
    int len = sizeof(error);    

    param = (ScanParam *)arg;
    epfd = param->epfd;
    limit = param->limit;
    timeout = param->timeout * 1000;
    events = malloc(limit * sizeof(struct epoll_event));
    memset(events, 0, limit * sizeof(struct epoll_event));

    //printf("wait_scanport limit\n", limit);

    while (1)
    {
        //printf("wait_scanport begin scan\n");

        pthread_mutex_lock(param->scan_mtx);
        while (param->bExit == 0 && param->bScan == 0) {
            //printf("wait_scanport wait scan\n");
            pthread_cond_wait(param->scan_cond, param->scan_mtx);
        }

        if (param->bExit == 1)
        {
            //printf("wait_scanport wait scan exit\n");
            pthread_mutex_unlock(param->scan_mtx);
            break;
        }

        param->bScan = 0;
        //printf("wait_scanport wait scan return timeout=%d\n", timeout);
        nfds = epoll_wait(epfd, events, limit, timeout);

        if (nfds == -1) 
        {
            //printf("epoll error.\n"); 
            break;    
        }

        for (i = 0; i < nfds; i++)
        {  
            scan = (ScanPort *)events[i].data.ptr;
            sockfd = scan->sockfd;

            if (events[i].events & EPOLLIN && events[i].events & EPOLLOUT)
            {
                //printf("Connect can read and write = 0\n");

                getsockopt(sockfd, SOL_SOCKET, SO_ERROR, &error, &len);
                
                if (error == 0)
                {
                    scan->result = 1; 
                    //printf("Connect success error = 0 port=%d\n", scan->port);
                }
                else
                {
                    //printf("Connect can read and write and Connect fail port=%d\n", scan->port); 
                }               
            }
            else if (events[i].events & EPOLLIN)
            {
                //printf("Connect fail port=%d\n", scan->port); 
            }
            else if (events[i].events & EPOLLOUT)
            {  
                scan->result = 1;    
                //printf("Connect success port=%d\n", scan->port);
            } 
            else
            {
                //printf("Connect other fail port=%d\n", scan->port);
            }
        }
        pthread_mutex_unlock(param->scan_mtx);

        //printf("wait_scanport wait begin.\n");
        pthread_mutex_lock(param->wait_mtx);
        param->bWait = 1;
        pthread_cond_signal(param->wait_cond);
        pthread_mutex_unlock(param->wait_mtx);
    }

    free(events);
}

static void initScan(ScanPort* scan, int count)
{
    int i;
    ScanPort* p;

    for (i = 0; i < count; i++)
    {
        p = scan + i;

        p->port = 0;
        memset(p->addr, 0, ADDR_LEN);

        p->result = 0;
        p->sockfd = -1;
    }
}

static void printHelp(char *prog)
{
    printf("%s usage.\n", prog);
    printf("eg:\n");
    printf("%s -s tcp -p 1-1024 -h 183.66.109.243 -l 1000 -t 3 -f fail.txt -S succ.txt\n", prog);
    printf("-s: scan method, value: tcp|syn\n");
    printf("-p: port range, value: 1-1024|22,80\n");
    printf("-h: remove address, value: 172.18.18.18\n");
    printf("-l: max sockets limit, value: 1000\n");
    printf("-l: socket connect timeout, value: 3\n");
    printf("-f: fail ip:port list, value: fail.txt\n");
    printf("-S: success ip:port list, value: success.txt\n");
}

static void parsePort(ScanParam *param, char *str)
{
    char *p = NULL;
    char *delim = "-";

    for (p = strtok(str, delim); p != NULL; p = strtok(NULL, delim))
    {
        if (param->start_port == 0)
        {
            param->start_port = atoi(p);
        }
        else
        {
            param->end_port = atoi(p);
        }
    }
 
    //printf("start port=%d\n", param->start_port);
    //printf("end port=%d\n", param->end_port);
}

/*
static void freeScan(ScanPort* scan, int count)
{
    int i;
    ScanPort* p;

    for (i = 0; i < count; i++)
    {
        p = scan + i;

        if (p->sockfd > 0)
        {
            close(p->sockfd);
            p->sockfd = -1;
        }
    }
}

static void write_result_file(ScanParam *param)
{
    int i = 0;
    FILE* pFile;
    
    ScanPort *p;
    ScanPort *scan = param->scan;

    pFile = fopen(param->failPath, "wb");
    if (pFile != NULL)  {
        for (i = 0; i < count; i++)
        {
            p = scan + i;

            if (p->result == 1)
            {
                continue;
            }

            memset(buf, 0, 128);
            sprintf(buf, "%s:%d\n", p->addr, p->port); 
            size = strlen(buf);

            fwrite(buf, (size_t)size, 1, pFile);
        }
        
        fclose( pFile );    pFile = NULL;
    }

    ////////////////////////////////////////
    pFile = fopen(param->succPath, "wb");
    if (pFile != NULL)  {  
        for (i = 0; i < count; i++)
        {
            p = scan + i;

            if (p->result == 0)
            {
                continue;
            }

            memset(buf, 0, 128);
            sprintf(buf, "%s:%d\n", p->addr, p->port); 
            size = strlen(buf);

            fwrite(buf, (size_t)size, 1, pFile);
        }

        fclose( pFile );    pFile = NULL;
    }
}
*/