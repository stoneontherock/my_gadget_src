/* Description: 面对黑盒想要知道某个操作对文件执行了什么操作(比如读/写/删除等)
 *              使用epoll多路复用，使用inotify()来监测事件
 * Author: zhouhui
 * Release Date: 2014-10-11
 * Modify log: 
 */ 

#include <stdio.h>
#include <string.h>
#include <unistd.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <stdlib.h>
#include <errno.h>
#include <time.h>
#include <dirent.h>
#include <sys/inotify.h>
#include <sys/epoll.h>

#define MAX_DIR_LEN 1024  /* max dir path length */
#define MAX_DIR_MON  4096 /* max dir monitored */ 
#define MAX_EPL_EVT_READY 64 /* max ready fd return from epoll_wait */ 
#define  MAXEVENTS 2048 /* argument 3 of epoll_wait() */ 

/* global variable */
int ifd; /*inotify fd*/
uint32_t msk = IN_DONT_FOLLOW | IN_EXCL_UNLINK; /*inotify event flag mask, do not follow the link file, exclude file be removed*/
char * wd_path[MAX_DIR_MON];
unsigned monitor_dir_cnt;

char path_str[MAX_DIR_MON * MAX_DIR_LEN];

/* function declaration*/
int dirTree(const char *dir);
int add_path_to_inotify(const char *path);
int get_inotify_mask(void);
void display_inotify_event(struct inotify_event *ie);
int read_ready_fd(int fd);

int main(int argc, char **argv)
{
    if (argc <2)
    {
        fprintf(stderr,"Usage: %s <directory_path>\n", argv[0]);
        return 0;
    }

    struct stat st;
    if (stat(argv[1], &st) == -1)
    {
        perror("stat");
        return 1;
    }

    //argument 1 must be a dir path
    if (!(st.st_mode & S_IFDIR))
    {
        fprintf(stderr, "Usage: %s <directory_path>\n" 
             "Agument 1 must be a directory path\n", argv[0]);
        return 1;
    }

    if (get_inotify_mask() == -1)
        return -1;

    ifd = inotify_init();
    if (ifd == -1)
    {
        perror("inotify_init");
        return 1;
    }

    if (dirTree(argv[1]) == -1)
        return -1;
    printf("%u directories being monitored...\n",monitor_dir_cnt);

    int epfd = epoll_create(1);
    if (epfd == -1)
    {
        perror("epoll_init");
        return 1;
    }

    // only one fd add to epoll. so epoll here not nessesary. 
    struct epoll_event ev;
    ev.data.fd = ifd;
    ev.events = EPOLLIN;
    if (epoll_ctl(epfd, EPOLL_CTL_ADD, ifd, &ev) == -1)
    {
        perror("epoll_ctl");
        return 1;
    }

    while(1)
    {
        struct epoll_event ev_list[MAX_EPL_EVT_READY];
        int ready;
        ready = epoll_wait(epfd, ev_list, MAXEVENTS, -1);
        if (ready == -1)
        {
            perror("epoll_wait");
            return 1;
        }

        int i;
        for (i = 0; i<ready; i++)
        {
            if (!(ev_list[i].events & EPOLLIN))
            {
                fprintf(stderr,"\033[1;31mERROR\033[0m:epoll_wait return ERROR event %d\n",ev_list[i].events);
                close(ev_list[i].data.fd);
            }
            if (read_ready_fd(ev_list[i].data.fd) == -1)
                return -1;
        }
    }

    return 0;
}

void display_inotify_event(struct inotify_event *ie)
{
    const time_t t = time(NULL);
    struct tm *tm = localtime(&t);  
    
    // move from and  move to display in one line.
    if (ie->mask & IN_MOVED_FROM)
    {
        printf("%02d:%02d:%02d  MOVED   : %s/%s",tm->tm_hour, tm->tm_min, tm->tm_sec ,wd_path[ie->wd], ie->name);
        return ;
    }
    if (ie->mask & IN_MOVED_TO) 
    {
        printf(" -> %s/%s\n",wd_path[ie->wd], ie->name);
        return ;
    }

    printf("%02d:%02d:%02d  ",tm->tm_hour, tm->tm_min, tm->tm_sec);
    if (ie->mask & IN_ACCESS)  printf("%-6s","ACCESS");
    if (ie->mask & IN_ATTRIB)  printf("%-6s","ATTRIB");
    if (ie->mask & IN_CREATE)  printf("%-6s","CREATE");
    if (ie->mask & IN_DELETE)  printf("%-6s","DELETE");
    if (ie->mask & IN_MODIFY)  printf("%-6s","MODIFY");
    printf("  : %s/%s\n",wd_path[ie->wd], ie->name);
}

int read_ready_fd(int fd)
{
    while (1)
    {
        char buf[2048] = {0};
        ssize_t readNum;
        readNum = read(fd, buf, sizeof(buf));
        if (readNum <= 0)
        {
            perror("read");
            return -1;
        }

        char * ptr = buf;
        while(ptr < buf + readNum)
        {
            struct inotify_event *ie = (struct inotify_event *)ptr;
            if (!(ie->mask & IN_ISDIR))
                display_inotify_event(ie); 
            ptr += sizeof(struct inotify_event) + ie->len; 
        }

/***************************************
        下面是第一次实现方式，但是发现有bug，将这段代码以注释形式保留下来 
        struct inotify_event *ie = (struct inotify_event *)buf;
        while((char *)ie < buf + readNum)
        {
            if (!(ie->mask & IN_ISDIR))
                display_inotify_event(ie); 
            //bug！ 下面直接+=会导致ie的地址增加了 (结构体长+len)*结构体长
            (char *)ie += sizeof(struct inotify_event) + ie->len; 
        }
***************************************/

    }
}

int get_inotify_mask(void)
{
    uint32_t msk_all[] = { IN_ACCESS, IN_ATTRIB, IN_MODIFY, IN_CREATE, IN_DELETE, IN_MOVE};
    char *prmt ="  1.\033[1;31mACCESS\033[0m : read() event\n"
                "  2.\033[1;31mATTRIB\033[0m : meta data changed event\n"
                "  3.\033[1;31mMODIFY\033[0m : modify file content event\n"
                "  4.\033[1;31mCREATE\033[0m : create file event\n"
                "  5.\033[1;31mDELETE\033[0m : delete file event\n"
                "  6.\033[1;31mMOVE\033[0m : move from here to there event\n"
                "which event do you want to monitor? [input numbers seperated by space] :";
    printf("%s", prmt);
    fflush(stdout);

    char buf[22];
    if (fgets(buf,sizeof(buf)-1, stdin) == NULL)
    {
        perror("fgets");
        return -1;
    }
    
    buf[strlen(buf) - 1] = 0;
    char *ptr, *str = buf;
    for(;;)
    {
        ptr = strtok(str," \t");
        if (ptr == NULL)
            break;
        int ind = atoi(ptr);
        if (ind <= 0 || ind > sizeof(msk_all)/sizeof(uint32_t))
        {
            fprintf(stderr,"\033[1;31mERROR\033[0m:invalid number\n");
            return -1;
        }
        msk |= msk_all[ind - 1];
        str = NULL;
   }

    return 0;
}

int add_path_to_inotify(const char *path)
{
    int wd = inotify_add_watch(ifd, path, msk);
    if (wd == -1)
    {
        fprintf(stderr,"\033[1;31mERROR\033[0m:<%s>\033[0m ",path);
        perror("inotify_add_watch");
        return -1;
    }

    static char *path_next = path_str;
    strncpy(path_next, path, strlen(path));
    wd_path[wd] = path_next;
    path_next += strlen(path)+1;
    if (path_str - path_next >= MAX_DIR_MON * MAX_DIR_LEN) 
    {
        fprintf(stderr,"\033[1;31mERROR\033[0m:too many directories\n");
        return -1;
    }
    monitor_dir_cnt++;

    return 0;
}

int dirTree(const char *dir)
{
    if (add_path_to_inotify(dir) == -1)
        return -1;

    DIR * dirp;
    dirp = opendir(dir);
    if (dirp == NULL)
    {
        fprintf(stderr,"\033[1;31mERROR\033[0m:%s",dir);
        perror("opendir");
        return -1;
    }
 
    struct dirent * dp; 
    while ( (dp = readdir(dirp)) )
    {
        errno = 0;
        if (dp == NULL)
        {
            if (errno == 0)
                break;
            perror("readdir");
            return -1;
        }

        if ( dp->d_name[0] == '.' || strcmp(dp->d_name,"..") == 0 )
            continue;

        if (dp->d_type != DT_DIR)
            continue;

        char path[MAX_DIR_LEN];
        sprintf(path,"%s/%s",dir,dp->d_name);
        if (dirTree(path) < 0)
        {
            return -1;
        }
    }

    return 0;
}
